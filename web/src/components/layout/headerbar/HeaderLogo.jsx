/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import React from 'react';
import { Link } from 'react-router-dom';

const HeaderLogo = ({
  systemName,
  logo,
  logoLoaded = true,
  showLogo = false,
  centered = false,
  linkTo = '/',
  className = '',
  textClassName = '',
  imageClassName = '',
}) => {
  const label = systemName || 'Sisyphus';
  const shouldShowLogo = Boolean(showLogo && logo && logoLoaded !== false);
  const wrapperClassName = [
    'group inline-flex items-center gap-3 transition-all duration-300',
    centered ? 'justify-center' : '',
    className,
  ]
    .filter(Boolean)
    .join(' ');
  const wordmarkClassName = [
    'text-2xl font-headline italic text-brand-primary transition-colors duration-300',
    linkTo ? 'group-hover:text-brand-primary-hover' : '',
    textClassName,
  ]
    .filter(Boolean)
    .join(' ');

  const content = (
    <>
      {shouldShowLogo && (
        <span
          className={[
            'flex h-11 w-11 items-center justify-center overflow-hidden rounded-full border border-brand-primary/15 bg-white shadow-sm',
            imageClassName,
          ]
            .filter(Boolean)
            .join(' ')}
        >
          <img src={logo} alt={`${label} logo`} className='h-full w-full object-cover' />
        </span>
      )}
      <span className={wordmarkClassName}>{label}</span>
    </>
  );

  if (!linkTo) {
    return <div className={wrapperClassName}>{content}</div>;
  }

  return (
    <Link to={linkTo} className={wrapperClassName}>
      {content}
    </Link>
  );
};

export default HeaderLogo;
