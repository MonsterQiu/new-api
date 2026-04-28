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

import React, { useContext, useEffect, useState } from 'react';
import { Button } from '@douyinfe/semi-ui';
import { API, showError, copy, showSuccess } from '../../helpers';
import { useIsMobile } from '../../hooks/common/useIsMobile';
import { StatusContext } from '../../context/Status';
import { useActualTheme } from '../../context/Theme';
import { marked } from 'marked';
import { useTranslation } from 'react-i18next';
import { IconCopy } from '@douyinfe/semi-icons';
import { Link } from 'react-router-dom';
import NoticeModal from '../../components/layout/NoticeModal';

const Home = () => {
  const { t, i18n } = useTranslation();
  const [statusState] = useContext(StatusContext);
  const actualTheme = useActualTheme();
  const [homePageContentLoaded, setHomePageContentLoaded] = useState(false);
  const [homePageContent, setHomePageContent] = useState('');
  const [noticeVisible, setNoticeVisible] = useState(false);
  const isMobile = useIsMobile();
  const serverAddress =
    statusState?.status?.server_address || `${window.location.origin}`;
  const isChinese = i18n.language.startsWith('zh');
  const gatewayBaseUrl = `${serverAddress.replace(/\/$/, '')}/v1`;
  const heroAnnouncement = statusState?.status?.version
    ? isChinese
      ? `Sisyphus ${statusState.status.version} 现已上线`
      : `Sisyphus ${statusState.status.version} is now live`
    : isChinese
      ? 'Sisyphus 网关现已就绪'
      : 'Sisyphus gateway is now ready';
  const heroTitle = isChinese ? 'Sisyphus API' : 'Sisyphus API';
  const heroAccent = isChinese
    ? '领体验额度QQ群：157708448'
    : 'Domestic Direct Access';
  const heroMetrics = isChinese
    ? [
        { value: '99.9%', label: '可用性' },
        { value: '<100ms', label: '路由延迟' },
        { value: '>98%', label: '缓存率' },
      ]
    : [
        { value: '99.9%', label: 'Uptime' },
        { value: '<100ms', label: 'Routing latency' },
        { value: '>98%', label: 'Cache rate' },
      ];
  const providerStrip = [
    { label: 'OPENAI', icon: 'psychiatry' },
    { label: 'ANTHROPIC', icon: 'neurology' },
    { label: 'GOOGLE', icon: 'token' },
    { label: 'META', icon: 'blur_on' },
    { label: 'COHERE', icon: 'data_object' },
  ];
  const protocolTags = ['OpenAI', 'Anthropic', 'Gemini', 'Llama 3'];
  const codeSample = [
    'from openai import OpenAI',
    '',
    '# Point the standard OpenAI client to Sisyphus',
    'client = OpenAI(',
    `    base_url="${gatewayBaseUrl}",`,
    '    api_key="sk-your-key-here"',
    ')',
    '',
    '# Call OpenAI',
    'gpt_res = client.chat.completions.create(',
    '    model="gpt-4-turbo",',
    '    messages=[{"role": "user", "content": "Hello"}]',
    ')',
    '',
    '# Call Claude using the exact same API',
    'claude_res = client.chat.completions.create(',
    '    model="claude-3-opus-20240229",',
    '    messages=[{"role": "user", "content": "Hello"}]',
    ')',
  ].join('\n');

  const displayHomePageContent = async () => {
    setHomePageContent(localStorage.getItem('home_page_content') || '');
    const res = await API.get('/api/home_page_content');
    const { success, message, data } = res.data;
    if (success) {
      let content = data;
      if (!data.startsWith('https://')) {
        content = marked.parse(data);
      }
      setHomePageContent(content);
      localStorage.setItem('home_page_content', content);

      // 如果内容是 URL，则发送主题模式
      if (data.startsWith('https://')) {
        const iframe = document.querySelector('iframe');
        if (iframe) {
          iframe.onload = () => {
            iframe.contentWindow.postMessage({ themeMode: actualTheme }, '*');
            iframe.contentWindow.postMessage({ lang: i18n.language }, '*');
          };
        }
      }
    } else {
      showError(message);
      setHomePageContent('加载首页内容失败...');
    }
    setHomePageContentLoaded(true);
  };

  const handleCopyBaseURL = async () => {
    const ok = await copy(serverAddress);
    if (ok) {
      showSuccess(t('已复制到剪切板'));
    }
  };

  const handleCopyCodeSample = async () => {
    const ok = await copy(codeSample);
    if (ok) {
      showSuccess(isChinese ? '代码示例已复制' : 'Code sample copied');
    }
  };

  useEffect(() => {
    const checkNoticeAndShow = async () => {
      const lastCloseDate = localStorage.getItem('notice_close_date');
      const today = new Date().toDateString();
      if (lastCloseDate !== today) {
        try {
          const res = await API.get('/api/notice');
          const { success, data } = res.data;
          if (success && data && data.trim() !== '') {
            setNoticeVisible(true);
          }
        } catch (error) {
          console.error('获取公告失败:', error);
        }
      }
    };

    checkNoticeAndShow();
  }, []);

  useEffect(() => {
    displayHomePageContent().then();
  }, []);

  return (
    <div className='w-full overflow-x-hidden'>
      <NoticeModal
        visible={noticeVisible}
        onClose={() => setNoticeVisible(false)}
        isMobile={isMobile}
      />
      {homePageContentLoaded && homePageContent === '' ? (
        <div className='w-full overflow-x-hidden'>
          {/* Banner 部分 */}
          <div className='w-full border-b border-semi-color-border min-h-[560px] md:min-h-[680px] lg:min-h-[760px] relative overflow-x-hidden'>
            {/* 背景模糊晕染球 */}
            {/* <div className='blur-ball blur-ball-indigo' />
            <div className='blur-ball blur-ball-teal' /> */}
            <div className='flex items-center justify-center h-full px-4 py-20 md:px-8 md:py-24 lg:py-32 mt-10'>
              {/* 居中内容区 */}
              <div className='flex flex-col items-center justify-center text-center max-w-6xl mx-auto w-full gap-8 md:gap-10'>
                <div className='inline-flex items-center gap-2 rounded-full border border-brand-primary/20 bg-white/70 px-4 py-2 text-sm font-body font-medium text-brand-primary shadow-sm backdrop-blur dark:bg-black/20'>
                  <span className='h-2 w-2 rounded-full bg-brand-primary animate-pulse' />
                  <span>{heroAnnouncement}</span>
                </div>

                <div className='flex flex-col items-center justify-center gap-4 md:gap-5'>
                  <h1
                    className={`font-headline text-3xl md:text-4xl lg:text-5xl xl:text-6xl font-bold text-semi-color-text-0 leading-[1.02] ${isChinese ? 'tracking-normal' : 'tracking-tight'}`}
                  >
                    {heroTitle}
                    <br />
                    <span className='text-brand-primary italic font-normal'>
                      {heroAccent}
                    </span>
                  </h1>

                  <div className='flex flex-wrap items-stretch justify-center gap-4 md:gap-6 w-full'>
                    {heroMetrics.map((metric) => (
                      <div
                        key={metric.label}
                        className='min-w-[150px] rounded-2xl border border-semi-color-border bg-white/60 px-5 py-4 shadow-sm backdrop-blur dark:bg-black/20'
                      >
                        <div className='font-mono text-xl md:text-2xl font-semibold text-semi-color-text-0'>
                          {metric.value}
                        </div>
                        <div className='mt-1 font-body text-sm text-semi-color-text-2'>
                          {metric.label}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>

                {/* BASE_URL + 主按钮 */}
                <div className='w-full max-w-[860px] flex flex-col items-center gap-3'>
                  <div className='w-full flex flex-col lg:flex-row items-stretch justify-center gap-4'>
                    <div className='flex min-w-0 flex-1 items-center rounded-2xl border border-[#2f2621] bg-[#171311] px-4 py-2 shadow-[inset_0_1px_0_rgba(255,255,255,0.03)]'>
                      <span className='mr-4 shrink-0 font-body text-xs tracking-[0.14em] text-[#8f7f74]'>
                        BASE_URL
                      </span>
                      <span className='min-w-0 flex-1 truncate font-mono text-sm md:text-base text-[#f7efe8]'>
                        {serverAddress}
                      </span>
                      <button
                        type='button'
                        onClick={handleCopyBaseURL}
                        className='ml-3 inline-flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-[#241d19] text-[#f7efe8] transition hover:bg-[#2d241f]'
                        aria-label={t('复制')}
                      >
                        <IconCopy />
                      </button>
                    </div>

                    <Link to='/console' className='shrink-0'>
                      <Button
                        theme='solid'
                        type='primary'
                        size={isMobile ? 'default' : 'large'}
                        className='!h-[52px] !rounded-2xl !border-0 !bg-[#c2652a] px-8 font-body text-base font-semibold shadow-[0_10px_30px_rgba(194,101,42,0.28)] hover:!bg-[#d9773b]'
                        icon={null}
                      >
                        {isChinese ? '获取 API Key' : 'Get API Key'}
                      </Button>
                    </Link>
                  </div>
                </div>
                <div
                  className='w-full max-w-[860px] rounded-2xl border border-[#c2652a]/70 bg-[#171311]/70 px-6 py-5 text-center
  shadow-[0_12px_36px_rgba(194,101,42,0.12)]'
                >
                  <div className='font-body text-sm font-semibold tracking-[0.08em] text-[#8f7f74]'>
                    联系方式
                  </div>
                  <div className='mt-2 font-headline text-2xl  text-[#c2652a] md:text-3xl'>
                    VX联系方式：sisyphusx_api
                  </div>
                </div>

                <section className='mt-8 w-full rounded-[32px] border border-[#241d19] bg-[#0b0908] px-4 py-6 text-left shadow-[0_30px_80px_rgba(18,12,9,0.32)] md:mt-12 md:px-6 md:py-8 lg:px-8 lg:py-10'>
                  <div className='relative overflow-hidden py-3'>
                    <div className='pointer-events-none absolute inset-y-0 left-0 w-12 bg-gradient-to-r from-[#0b0908] to-transparent md:w-20' />
                    <div className='pointer-events-none absolute inset-y-0 right-0 w-12 bg-gradient-to-l from-[#0b0908] to-transparent md:w-20' />
                    <div className='sisyphus-provider-marquee flex w-max items-center gap-10 pr-10 font-body text-[11px] font-semibold uppercase tracking-[0.24em] text-[#8f7f74] md:text-xs'>
                      {[...providerStrip, ...providerStrip].map(
                        (provider, index) => (
                          <div
                            key={`${provider.label}-${index}`}
                            className='flex items-center gap-2 whitespace-nowrap'
                          >
                            <span className='material-symbols-outlined text-[16px]'>
                              {provider.icon}
                            </span>
                            <span>{provider.label}</span>
                          </div>
                        ),
                      )}
                    </div>
                  </div>

                  <div className='mt-6 overflow-hidden rounded-[24px] border border-[#241d19] bg-[#11100f] shadow-[0_20px_60px_rgba(0,0,0,0.24)] md:mt-8'>
                    <div className='flex items-center border-b border-[#241d19] bg-[#1a1613] px-4 py-3'>
                      <div className='flex gap-2'>
                        <span className='h-3 w-3 rounded-full bg-[#4a433d]' />
                        <span className='h-3 w-3 rounded-full bg-[#4a433d]' />
                        <span className='h-3 w-3 rounded-full bg-[#4a433d]' />
                      </div>
                      <span className='mx-auto font-mono text-xs text-[#8f7f74]'>
                        multi_model.py
                      </span>
                      <button
                        type='button'
                        onClick={handleCopyCodeSample}
                        className='text-[#8f7f74] transition hover:text-[#f7efe8]'
                        aria-label={
                          isChinese ? '复制代码示例' : 'Copy code sample'
                        }
                      >
                        <IconCopy />
                      </button>
                    </div>

                    <pre className='overflow-x-auto px-5 py-6 font-mono text-sm leading-7 text-[#d6d3d1] md:px-6 md:text-[15px]'>
                      <code>
                        <span className='text-[#fb923c]'>from</span> openai{' '}
                        <span className='text-[#fb923c]'>import</span> OpenAI
                        {'\n\n'}
                        <span className='text-[#78716c]'>
                          {isChinese
                            ? '# 将标准 OpenAI 客户端指向 Sisyphus'
                            : '# Point the standard OpenAI client to Sisyphus'}
                        </span>
                        {'\n'}
                        client = OpenAI({'\n'}
                        {'    '}base_url=
                        <span className='text-[#4ade80]'>
                          "{gatewayBaseUrl}"
                        </span>
                        ,{'\n'}
                        {'    '}api_key=
                        <span className='text-[#4ade80]'>
                          "sk-your-key-here"
                        </span>
                        {'\n'}){'\n\n'}
                        <span className='text-[#78716c]'>
                          {isChinese ? '# 调用 OpenAI' : '# Call OpenAI'}
                        </span>
                        {'\n'}
                        gpt_res = client.chat.completions.create({'\n'}
                        {'    '}model=
                        <span className='text-[#4ade80]'>"gpt-4-turbo"</span>,
                        {'\n'}
                        {'    '}messages=[{'{'}
                        <span className='text-[#4ade80]'>"role"</span>:{' '}
                        <span className='text-[#4ade80]'>"user"</span>,{' '}
                        <span className='text-[#4ade80]'>"content"</span>:{' '}
                        <span className='text-[#4ade80]'>"Hello"</span>
                        {'}'}]{'\n'}){'\n\n'}
                        <span className='text-[#78716c]'>
                          {isChinese
                            ? '# 用同一套 API 调用 Claude'
                            : '# Call Claude using the exact same API'}
                        </span>
                        {'\n'}
                        claude_res = client.chat.completions.create({'\n'}
                        {'    '}model=
                        <span className='text-[#4ade80]'>
                          "claude-3-opus-20240229"
                        </span>
                        ,{'\n'}
                        {'    '}messages=[{'{'}
                        <span className='text-[#4ade80]'>"role"</span>:{' '}
                        <span className='text-[#4ade80]'>"user"</span>,{' '}
                        <span className='text-[#4ade80]'>"content"</span>:{' '}
                        <span className='text-[#4ade80]'>"Hello"</span>
                        {'}'}]{'\n'})
                      </code>
                    </pre>
                  </div>

                  <div className='mt-6 grid gap-4 md:mt-8 md:grid-cols-3 md:gap-6'>
                    <article className='relative overflow-hidden rounded-[24px] border border-[#241d19] bg-[#1a1512] p-6 transition-colors hover:border-brand-primary/30 md:col-span-2 md:p-8'>
                      <div className='absolute right-0 top-0 h-44 w-44 translate-x-1/4 -translate-y-1/3 rounded-full bg-brand-primary/10 blur-3xl' />
                      <h3 className='relative font-headline text-3xl text-[#f7efe8]'>
                        {isChinese ? '统一协议' : 'Unified Protocol'}
                      </h3>
                      <p className='relative mt-3 max-w-2xl font-body text-base leading-7 text-[#a8a29e]'>
                        {isChinese
                          ? '一切都说 OpenAI 格式。继续使用现有的 OpenAI SDK、LangChain 或 LlamaIndex 接入任意模型，无需重写业务代码。'
                          : 'Everything speaks OpenAI format. Use your existing OpenAI SDKs, LangChain, or LlamaIndex integrations to talk to any model. No code rewrites needed.'}
                      </p>
                      <div className='relative mt-6 flex flex-wrap gap-2'>
                        {protocolTags.map((tag) => (
                          <span
                            key={tag}
                            className='rounded-md border border-[#3a312c] bg-[#14110f] px-3 py-1 font-mono text-xs text-[#d6d3d1]'
                          >
                            {tag}
                          </span>
                        ))}
                      </div>
                    </article>

                    <article className='flex flex-col justify-between rounded-[24px] border border-[#241d19] bg-[#171311] p-6 transition-colors hover:border-brand-primary/30 md:p-8'>
                      <div>
                        <div className='mb-4 inline-flex h-10 w-10 items-center justify-center rounded-xl bg-brand-primary/10 font-mono text-xs font-semibold uppercase tracking-[0.2em] text-brand-primary'>
                          RT
                        </div>
                        <h3 className='font-headline text-2xl text-[#f7efe8]'>
                          {isChinese ? '智能路由' : 'Smart Routing'}
                        </h3>
                        <p className='mt-3 font-body text-sm leading-6 text-[#a8a29e]'>
                          {isChinese
                            ? '在上游故障或延迟波动时，自动切换到回退供应商，保持网关稳定输出。'
                            : 'Auto-switch to fallback providers on downtime or latency spikes.'}
                        </p>
                      </div>

                      <div className='mt-8 space-y-3'>
                        <div className='flex h-1.5 overflow-hidden rounded-full bg-[#2a241f]'>
                          <div className='h-full w-[60%] bg-brand-primary' />
                          <div className='h-full w-[40%] bg-[#4a433d]' />
                        </div>
                        <div className='flex h-1.5 overflow-hidden rounded-full bg-[#2a241f]'>
                          <div className='h-full w-[30%] bg-[#4a433d]' />
                          <div className='h-full w-[70%] bg-brand-primary' />
                        </div>
                      </div>
                    </article>

                    <article className='rounded-[24px] border border-[#241d19] bg-[#171311] p-6 transition-colors hover:border-brand-primary/30 md:p-8'>
                      <div className='mb-4 inline-flex h-10 w-10 items-center justify-center rounded-xl bg-brand-primary/10 font-mono text-xs font-semibold uppercase tracking-[0.2em] text-brand-primary'>
                        CC
                      </div>
                      <h3 className='font-headline text-2xl text-[#f7efe8]'>
                        {isChinese ? '成本控制' : 'Cost Control'}
                      </h3>
                      <p className='mt-3 font-body text-sm leading-6 text-[#a8a29e]'>
                        {isChinese
                          ? '为用户设置额度，追踪 token 使用量，并在同一面板中观察所有模型的成本走势。'
                          : 'Set per-user quotas, monitor token usage, and track spend across all your models in one dashboard.'}
                      </p>
                    </article>

                    <article className='rounded-[24px] border border-[#241d19] bg-[#171311] p-6 transition-colors hover:border-brand-primary/30 md:col-span-2 md:p-8'>
                      <div className='mb-6 flex flex-col gap-4 md:flex-row md:items-start md:justify-between'>
                        <div>
                          <div className='mb-4 inline-flex h-10 w-10 items-center justify-center rounded-xl bg-brand-primary/10 font-mono text-xs font-semibold uppercase tracking-[0.2em] text-brand-primary'>
                            SC
                          </div>
                          <h3 className='font-headline text-2xl text-[#f7efe8]'>
                            {isChinese ? '语义缓存' : 'Semantic Caching'}
                          </h3>
                        </div>
                        <div className='inline-flex items-center rounded-full border border-[#24523f] bg-[#0f1d17] px-3 py-1 text-xs font-medium text-[#86efac]'>
                          {isChinese
                            ? '99.9% 缓存命中率'
                            : '99.9% Cache Hit Ratio'}
                        </div>
                      </div>

                      <div className='relative h-24 overflow-hidden rounded-xl border border-[#241d19] bg-[#0f0b09]'>
                        <div className='absolute bottom-0 left-0 h-1/2 w-full bg-gradient-to-t from-brand-primary/10 to-transparent' />
                        <svg
                          className='absolute inset-0 h-full w-full'
                          preserveAspectRatio='none'
                          viewBox='0 0 100 100'
                        >
                          <path
                            d='M0,100 L0,50 Q25,60 50,30 T100,10 L100,100 Z'
                            fill='none'
                            stroke='#c2652a'
                            strokeWidth='2'
                            className='opacity-80'
                          />
                        </svg>
                      </div>
                    </article>
                  </div>
                </section>
              </div>
            </div>
          </div>
        </div>
      ) : (
        <div className='overflow-x-hidden w-full'>
          {homePageContent.startsWith('https://') ? (
            <iframe
              src={homePageContent}
              className='w-full h-screen border-none'
            />
          ) : (
            <div
              className='mt-[60px]'
              dangerouslySetInnerHTML={{ __html: homePageContent }}
            />
          )}
        </div>
      )}
    </div>
  );
};

export default Home;
