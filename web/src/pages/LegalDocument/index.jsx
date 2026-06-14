import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import { Button, Card, Empty, Typography } from '@douyinfe/semi-ui';
import MarkdownRenderer from '../../components/common/markdown/MarkdownRenderer';
import { LEGAL_DOCUMENTS } from '../../constants/legalDocuments';

const { Title, Text } = Typography;

const buildContent = (document) => `## ${document.title}

${document.summary}

### 使用要求

- 你应当在访问、注册、登录或调用本服务前阅读并理解本页面对应条款。
- 你不得将本服务用于违法违规、滥用、攻击、批量注册、转售接口、规避地区限制或其他破坏性行为。
- 如官方条款、政策或支持地区发生更新，以官方页面的最新内容为准。
- 继续使用本服务即表示你确认已阅读、理解并同意遵守相关条款。

### 官方内容

[查看${document.title}](${document.externalUrl})
`;

const LegalDocument = () => {
  const location = useLocation();
  const documentKey = location.pathname.replace(/^\/+/, '');
  const document = LEGAL_DOCUMENTS.find((item) => item.key === documentKey);

  if (!document) {
    return (
      <div className='flex justify-center items-center min-h-screen bg-gray-50 p-4'>
        <Empty title='未找到该文档' />
      </div>
    );
  }

  return (
    <div className='min-h-screen bg-gray-50'>
      <div className='max-w-4xl mx-auto py-12 px-4 sm:px-6 lg:px-8'>
        <Card className='!rounded-lg'>
          <div className='mb-6'>
            <Link to='/login' className='text-blue-600 hover:text-blue-800'>
              返回登录
            </Link>
          </div>
          <Title heading={2} className='text-center mb-4'>
            {document.title}
          </Title>
          <Text type='tertiary' className='block text-center mb-8'>
            请阅读本页说明，并通过官方链接查看完整内容。
          </Text>
          <div className='prose prose-lg max-w-none'>
            <MarkdownRenderer content={buildContent(document)} />
          </div>
          <div className='mt-8 text-center'>
            <Button
              theme='solid'
              type='primary'
              onClick={() => window.open(document.externalUrl, '_blank')}
            >
              打开官方{document.title}
            </Button>
          </div>
        </Card>
      </div>
    </div>
  );
};

export default LegalDocument;
