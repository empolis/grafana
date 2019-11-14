import React, { FC } from 'react';
import { Tooltip } from '@grafana/ui';

interface Props {
  appName: string;
  buildVersion: string;
  buildCommit: string;
  newGrafanaVersionExists: boolean;
  newGrafanaVersion: string;
}

export const Footer: FC<Props> = React.memo(
  ({ appName, buildVersion, buildCommit, newGrafanaVersionExists, newGrafanaVersion }) => {
    return (
      <footer className="footer">
        <div className="text-center">
          <ul>
            <li>
              <a href="https://empolis.com/empolis-industrial-analytics/" target="_blank" rel="noopener">
                Empolis Industrial Analytics
              </a>{' '}
              <span>
                v{buildVersion} (commit: {buildCommit})
              </span>
            </li>
          </ul>
        </div>
      </footer>
    );
  }
);

export default Footer;
