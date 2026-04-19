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
import { Spin } from '@douyinfe/semi-ui';

const Loading = ({ size = 'small' }) => {
  return (
    <div className='fixed inset-0 z-[200] flex h-screen w-screen items-center justify-center overflow-hidden bg-[var(--app-shell-bg)] text-[var(--semi-color-text-0)]'>
      <div className='absolute inset-0 bg-[radial-gradient(circle_at_18%_12%,rgba(194,101,42,0.18),transparent_24%),radial-gradient(circle_at_82%_16%,rgba(255,237,213,0.78),transparent_22%),linear-gradient(180deg,#fbf7f2_0%,#f5eee6_52%,#fbf8f4_100%)] dark:bg-[radial-gradient(circle_at_18%_12%,rgba(194,101,42,0.18),transparent_26%),radial-gradient(circle_at_82%_16%,rgba(120,53,15,0.18),transparent_22%),linear-gradient(180deg,#100c0a_0%,#15100d_52%,#0f0b09_100%)]' />
      <div className='relative mx-4 w-full max-w-xl rounded-[28px] border border-[rgba(138,99,73,0.18)] bg-[rgba(255,251,247,0.72)] px-8 py-10 text-center shadow-[0_24px_72px_rgba(74,45,24,0.14)] backdrop-blur-xl dark:border-[rgba(251,232,216,0.1)] dark:bg-[rgba(21,17,15,0.82)] dark:shadow-[0_24px_72px_rgba(0,0,0,0.36)] md:px-12 md:py-12'>
        <div className='inline-flex h-14 w-14 items-center justify-center rounded-2xl bg-[rgba(194,101,42,0.12)] text-brand-primary shadow-[inset_0_1px_0_rgba(255,255,255,0.25)] dark:bg-[rgba(194,101,42,0.16)]'>
          <Spin size={size} spinning={true} />
        </div>

        <div className='mt-6'>
          <div className='font-headline text-4xl italic tracking-tight text-brand-primary md:text-5xl'>
            Sisyphus
          </div>
          <p className='mt-4 font-body text-base leading-7 text-[var(--semi-color-text-1)] md:text-lg'>
            代码即巨石，协议为山径。
          </p>
          <p className='mt-2 font-body text-sm tracking-[0.18em] text-[var(--semi-color-text-2)] uppercase'>
            Routing models through one path
          </p>
        </div>

        <div className='mt-8 overflow-hidden rounded-full bg-[rgba(36,29,25,0.08)] p-1 dark:bg-[rgba(255,255,255,0.06)]'>
          <div className='loading-progress-bar h-1.5 rounded-full bg-brand-primary/80' />
        </div>
      </div>
    </div>
  );
};

export default Loading;
