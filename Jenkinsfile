@Library('ese-ia-deployment-config') _

pipeline {
    agent {
        label 'docker'
    }

    environment {
        IMAGE = 'eseia/grafana'
        GIT_COMMIT_SHORT = iaGitCommitShort()
        GIT_BRANCH_TAG = GIT_BRANCH.replace('/', '_')
        COMMON_BUILD_ARGS = """\
            --label org.opencontainers.image.vendor="Empolis Information Management GmbH" \
            --label org.opencontainers.image.title=$IMAGE \
            --label org.opencontainers.image.created="$BUILD_TIMESTAMP_RFC3339" \
            --label org.opencontainers.image.revision=$GIT_BRANCH_TAG-$GIT_COMMIT_SHORT \
            .""".stripIndent()
    }

    stages {
        stage('Prepare workspace') {
            steps {
                sh "/bin/echo -n $GIT_BRANCH > git-branch"
                sh "/bin/echo -n $GIT_COMMIT_SHORT > git-sha"
                sh "/bin/echo -n \$(git show -s --format=%ct) > git-buildstamp"
            }
        }

        stage('Ensure builder is running') {
            steps {
                iaBuildxBootstrap()
            }
        }

        stage('Build and push') {
            steps {
                iaBuildxBuild COMMON_BUILD_ARGS
            }
        }
    }
}
