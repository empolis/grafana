import React, { FC } from 'react';
import { css, cx } from 'emotion';
import config from 'app/core/config';
import { useTheme } from '@grafana/ui';

export interface BrandComponentProps {
  className?: string;
  children?: JSX.Element | JSX.Element[];
}

const LoginLogo: FC<BrandComponentProps> = ({ className }) => {
  const { empolisOptions } = config;
  if (empolisOptions.customLogo.trim()) {
    return <img className={className} src={empolisOptions.customLogo} alt="Empolis" />;
  }
  return null;
};

const LoginBackground: FC<BrandComponentProps> = ({ className, children }) => {
  const { empolisOptions } = config;
  const background = css`
    background: url(${empolisOptions.loginBgImg});
    background-size: cover;
  `;

  return <div className={cx(background, className)}>{children}</div>;
};

const MenuLogo: FC<BrandComponentProps> = ({ className }) => {
  const { empolisOptions } = config;
  return <img className={className} src={empolisOptions.menuLogo} alt="Empolis" />;
};

const LoginBoxBackground = () => {
  const theme = useTheme();
  return css`
    background: ${theme.isLight ? 'rgba(6, 30, 200, 0.1 )' : 'rgba(18, 28, 41, 0.65)'};
    background-size: cover;
  `;
};

export class Branding {
  static LoginLogo = LoginLogo;
  static LoginBackground = LoginBackground;
  static MenuLogo = MenuLogo;
  static LoginBoxBackground = LoginBoxBackground;
  static AppTitle = 'Grafana';
  static LoginTitle = 'Welcome to Grafana';
  static GetLoginSubTitle = () => {
    const slogans = [
      "Don't get in the way of the data",
      'Your single pane of glass',
      'Built better together',
      'Democratising data',
    ];
    const count = slogans.length;
    return slogans[Math.floor(Math.random() * count)];
  };
}
