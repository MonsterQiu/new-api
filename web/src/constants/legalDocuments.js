export const LEGAL_DOCUMENTS = [
  {
    key: 'terms-of-service',
    title: '服务条款',
    path: '/terms-of-service',
    externalUrl: 'https://openai.com/policies/terms-of-use/',
    summary:
      '使用本服务即表示你同意遵守平台账号、计费、访问、内容与服务使用相关条款。',
  },
  {
    key: 'usage-policy',
    title: '使用政策',
    path: '/usage-policy',
    externalUrl: 'https://openai.com/policies/usage-policies/',
    summary:
      '不得使用本服务生成、传播或协助违法、欺诈、滥用、骚扰、规避安全机制等违规内容或行为。',
  },
  {
    key: 'supported-countries',
    title: '支持的国家和地区',
    path: '/supported-countries',
    externalUrl:
      'https://help.openai.com/en/articles/5347006-openai-api-supported-countries-and-territories',
    summary:
      '你需要确认自己位于服务支持的国家或地区，且不会通过规避地区限制的方式访问服务。',
  },
  {
    key: 'service-terms',
    title: '服务特定条款',
    path: '/service-terms',
    externalUrl: 'https://openai.com/policies/service-terms/',
    summary:
      '部分模型、能力和第三方服务可能适用额外条款，使用前需要阅读并遵守对应服务特定条款。',
  },
];

export const LEGAL_CONSENT_MESSAGE =
  '请先阅读并同意服务条款、使用政策、支持的国家和地区、服务特定条款';
