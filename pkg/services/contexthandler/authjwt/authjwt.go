package authjwt

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"strings"
	"time"

	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/infra/remotecache"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/sqlstore"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/grafana/grafana/pkg/util/errutil"
	"github.com/jmespath/go-jmespath"
)

const CachePrefix = "auth-jwt:sync-%s"

type AuthJWT struct {
	cfg            *setting.Cfg
	jwtToken       string
	jwtAuthService models.JWTService
	remoteCache    *remotecache.RemoteCache
	ctx            *models.ReqContext
	orgID          int64
}

type Error struct {
	Message      string
	DetailsError error
}

func newError(message string, err error) Error {
	return Error{
		Message:      message,
		DetailsError: err,
	}
}

func (err Error) Error() string {
	return err.Message
}

type Options struct {
	JWTAuthService models.JWTService
	RemoteCache    *remotecache.RemoteCache
	Ctx            *models.ReqContext
	OrgID          int64
}

func New(cfg *setting.Cfg, jwtToken string, options *Options) *AuthJWT {
	return &AuthJWT{
		cfg:            cfg,
		jwtToken:       jwtToken,
		jwtAuthService: options.JWTAuthService,
		remoteCache:    options.RemoteCache,
		ctx:            options.Ctx,
		orgID:          options.OrgID,
	}
}

func HashCacheKey(key string) (string, error) {
	hasher := fnv.New128a()
	if _, err := hasher.Write([]byte(key)); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func (auth *AuthJWT) getKey() (string, error) {
	hashedKey, err := HashCacheKey(auth.jwtToken)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(CachePrefix, hashedKey), nil
}

func (auth *AuthJWT) HandleError(err error, statusCode int, cb func(error)) {
	details := err
	var e Error
	if errors.As(err, &e) {
		details = e.DetailsError
	}
	auth.ctx.Handle(auth.cfg, statusCode, err.Error(), details)

	if cb != nil {
		cb(details)
	}
}

func (auth *AuthJWT) Login(logger log.Logger, ignoreCache bool) (int64, error) {
	if !ignoreCache {
		// Error here means absent cache - we don't need to handle that
		id, err := auth.GetUserViaCache(logger)
		if err == nil && id != 0 {
			return id, nil
		}
	}

	id, err := auth.LoginViaJWT(logger)
	if err != nil {
		return 0, newError("failed to log in as user, specified in jwt", err)
	}

	return id, nil
}

func (auth *AuthJWT) GetUserViaCache(logger log.Logger) (int64, error) {
	cacheKey, err := auth.getKey()
	if err != nil {
		return 0, err
	}
	logger.Debug("Getting user ID via auth cache", "cacheKey", cacheKey)
	userID, err := auth.remoteCache.Get(cacheKey)
	if err != nil {
		logger.Debug("Failed getting user ID via auth cache", "error", err)
		return 0, err
	}

	logger.Debug("Successfully got user ID via auth cache", "id", userID)
	return userID.(int64), nil
}

func (auth *AuthJWT) RemoveUserFromCache(logger log.Logger) error {
	cacheKey, err := auth.getKey()
	if err != nil {
		return err
	}
	logger.Debug("Removing user from auth cache", "cacheKey", cacheKey)
	if err := auth.remoteCache.Delete(cacheKey); err != nil {
		return err
	}

	logger.Debug("Successfully removed user from auth cache", "cacheKey", cacheKey)
	return nil
}

func (auth *AuthJWT) LoginViaJWT(logger log.Logger) (int64, error) {
	claims, err := auth.jwtAuthService.Verify(auth.ctx.Req.Context(), auth.jwtToken)
	if err != nil {
		auth.HandleError(err, 401, func(details error) {
			logger.Error(
				"Failed to verify JWT",
				"message", err.Error(),
				"error", details,
			)
		})
		return 0, fmt.Errorf("Failed to verify JWT")
	}

	if claims["sub"] == nil {
		return 0, fmt.Errorf("Failed to get an authentication claim from JWT")
	}

	extUser := &models.ExternalUserInfo{
		AuthModule: "auth_jwt",
		AuthId:     claims["sub"].(string),
		OrgRoles:   map[int64]models.RoleType{},
	}

	if key := auth.cfg.JWTAuthUsernameClaim; key != "" {
		extUser.Login, _ = claims[key].(string)
	}
	if key := auth.cfg.JWTAuthEmailClaim; key != "" {
		extUser.Email, _ = claims[key].(string)
	}
	if key := auth.cfg.JWTAuthNameClaim; key != "" {
		extUser.Name, _ = claims[key].(string)
	}

	if extUser.Login == "" && extUser.Email == "" {
		return 0, fmt.Errorf("Failed to get an authentication claim from JWT")
	}

	jwt, err := json.Marshal(claims)
	if err != nil {
		auth.HandleError(err, 401, func(details error) {
			logger.Error(
				"Failed to get JSON from JWT",
				"message", err.Error(),
				"error", details,
			)
		})
		return 0, fmt.Errorf("Failed to get JSON from JWT")
	}

	role, err := auth.extractRole(jwt)
	if err != nil {
		auth.HandleError(err, 401, func(details error) {
			logger.Error(
				"Failed to extract role from JWT",
				"message", err.Error(),
				"error", details,
			)
		})
		return 0, fmt.Errorf("Failed to extract role from JWT")
	} else {
		logger.Debug("Extracted role from JWT", "role", role)
	}

	var groups []string
	if auth.cfg.JWTGroupsAttributePath != "" {
		groups, err = auth.searchJSONForStringArrayAttr(auth.cfg.JWTGroupsAttributePath, jwt)
		if err != nil {
			auth.HandleError(err, 401, func(details error) {
				logger.Error(
					"Failed to extract groups from JWT",
					"message", err.Error(),
					"error", details,
				)
			})
			return 0, fmt.Errorf("Failed to extract groups from JWT")
		} else {
			logger.Debug("Extracted groups from JWT", "groups", groups)
		}
	}

	if role != "" {
		rt := models.RoleType(role)
		if rt.IsValid() {
			// The user will be assigned a role in either the auto-assigned organization or in the default one
			var orgID int64
			if setting.AutoAssignOrg && setting.AutoAssignOrgId > 0 {
				orgID = int64(setting.AutoAssignOrgId)
				logger.Debug("The user has a role assignment and organization membership is auto-assigned",
					"role", role, "orgId", orgID)
				extUser.OrgRoles[orgID] = rt
			} else if len(groups) > 0 {
				for _, group := range groups {
					query := models.GetOrgByNameQuery{Name: group}
					err := sqlstore.GetOrgByName(&query)
					if err != nil {
						query = models.GetOrgByNameQuery{Name: strings.ToLower(group)}
						err = sqlstore.GetOrgByName(&query)
					}
					if err == nil {
						extUser.OrgRoles[query.Result.Id] = rt
					}
				}
			} else {
				orgID = int64(1)
				logger.Debug("The user has a role assignment and organization membership is not auto-assigned",
					"role", role, "orgId", orgID)
				extUser.OrgRoles[orgID] = rt
			}
			logger.Debug("Set org roles", "extUser.OrgRoles", extUser.OrgRoles)
		}
	}

	upsert := &models.UpsertUserCommand{
		ReqContext:    auth.ctx,
		SignupAllowed: auth.cfg.JWTAllowSignup,
		ExternalUser:  extUser,
	}

	err = bus.Dispatch(upsert)
	if err != nil {
		return 0, err
	}

	return upsert.Result.Id, nil
}

func (auth *AuthJWT) GetSignedInUser(userID int64) (*models.SignedInUser, error) {
	query := &models.GetSignedInUserQuery{
		OrgId:  auth.orgID,
		UserId: userID,
	}

	if err := bus.DispatchCtx(context.Background(), query); err != nil {
		return nil, err
	}

	return query.Result, nil
}

func (auth *AuthJWT) Remember(id int64) error {
	cacheKey, err := auth.getKey()
	if err != nil {
		return err
	}

	// Check if user is already in cache
	userID, err := auth.remoteCache.Get(cacheKey)
	if err == nil && userID != nil {
		return nil
	}

	expiration := time.Duration(auth.cfg.JWTAuthSyncTTL) * time.Minute

	if err := auth.remoteCache.Set(cacheKey, id, expiration); err != nil {
		return err
	}

	return nil
}

func (auth *AuthJWT) extractRole(data []byte) (string, error) {
	if auth.cfg.JWTRoleAttributePath == "" {
		return "", nil
	}

	role, err := auth.searchJSONForStringAttr(auth.cfg.JWTRoleAttributePath, data)
	if err != nil {
		return "", err
	}
	return role, nil
}

func (auth *AuthJWT) searchJSONForStringAttr(attributePath string, data []byte) (string, error) {
	val, err := auth.searchJSONForAttr(attributePath, data)
	if err != nil {
		return "", err
	}

	strVal, ok := val.(string)
	if ok {
		return strVal, nil
	}

	return "", nil
}

func (auth *AuthJWT) searchJSONForAttr(attributePath string, data []byte) (interface{}, error) {
	if attributePath == "" {
		return "", errors.New("no attribute path specified")
	}

	if len(data) == 0 {
		return "", errors.New("empty JWT provided")
	}

	var buf interface{}
	if err := json.Unmarshal(data, &buf); err != nil {
		return "", errutil.Wrap("failed to unmarshal JWT", err)
	}

	val, err := jmespath.Search(attributePath, buf)
	if err != nil {
		return "", errutil.Wrapf(err, "failed to search JWT with provided path: %q", attributePath)
	}

	return val, nil
}

func (auth *AuthJWT) searchJSONForStringArrayAttr(attributePath string, data []byte) ([]string, error) {
	val, err := auth.searchJSONForAttr(attributePath, data)
	if err != nil {
		return []string{}, err
	}

	ifArr, ok := val.([]interface{})
	if !ok {
		return []string{}, nil
	}

	result := []string{}
	for _, v := range ifArr {
		if strVal, ok := v.(string); ok {
			result = append(result, strVal)
		}
	}

	return result, nil
}
