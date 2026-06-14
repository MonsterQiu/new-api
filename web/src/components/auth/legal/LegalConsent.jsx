import React from 'react';
import { Checkbox } from '@douyinfe/semi-ui';
import Text from '@douyinfe/semi-ui/lib/es/typography/text';
import { LEGAL_DOCUMENTS } from '../../../constants/legalDocuments';

const LegalConsent = ({ checked, onChange, className = '' }) => {
  return (
    <div className={className}>
      <Checkbox checked={checked} onChange={(e) => onChange(e.target.checked)}>
        <Text size='small' className='text-gray-600 leading-6'>
          我已阅读并同意
          {LEGAL_DOCUMENTS.map((document, index) => (
            <React.Fragment key={document.key}>
              {index > 0 && (index === LEGAL_DOCUMENTS.length - 1 ? '、' : '、')}
              <a
                href={document.path}
                target='_blank'
                rel='noopener noreferrer'
                className='text-blue-600 hover:text-blue-800 mx-1'
              >
                {document.title}
              </a>
            </React.Fragment>
          ))}
        </Text>
      </Checkbox>
    </div>
  );
};

export default LegalConsent;
