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

import React, { useContext, useEffect, useMemo, useRef, useState } from 'react';
import { Card, Empty, Table, Tag, Typography } from '@douyinfe/semi-ui';
import {
  IllustrationNoResult,
  IllustrationNoResultDark,
} from '@douyinfe/semi-illustrations';
import { useTranslation } from 'react-i18next';
import { UserContext } from '../../context/User';
import { StatusContext } from '../../context/Status';
import {
  API,
  copy,
  getQuotaPerUnit,
  renderQuota,
  showError,
  showSuccess,
  timestamp2string,
} from '../../helpers';
import InvitationCard from '../../components/topup/InvitationCard';
import TransferModal from '../../components/topup/modals/TransferModal';

const { Text } = Typography;

const Aff = () => {
  const { t } = useTranslation();
  const [userState, userDispatch] = useContext(UserContext);
  const [statusState] = useContext(StatusContext);
  const [affLink, setAffLink] = useState('');
  const [openTransfer, setOpenTransfer] = useState(false);
  const [transferAmount, setTransferAmount] = useState(0);
  const [rebates, setRebates] = useState([]);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [total, setTotal] = useState(0);
  const affFetchedRef = useRef(false);
  const inviteRebateRatio = Number(statusState?.status?.invite_rebate_ratio);
  const hasInviteRebateRatio =
    Number.isFinite(inviteRebateRatio) && inviteRebateRatio > 0;
  const inviteRebateEnabled =
    !!statusState?.status?.invite_rebate_enabled && hasInviteRebateRatio;
  const inviteRebatePercent = hasInviteRebateRatio
    ? `${(inviteRebateRatio * 100).toFixed(2).replace(/\.?0+$/, '')}%`
    : '0%';

  const getUserQuota = async () => {
    const res = await API.get('/api/user/self');
    const { success, message, data } = res.data;
    if (success) {
      userDispatch({ type: 'login', payload: data });
    } else {
      showError(message);
    }
  };

  const getAffLink = async () => {
    const res = await API.get('/api/user/aff');
    const { success, message, data } = res.data;
    if (success) {
      setAffLink(`${window.location.origin}/register?aff=${data}`);
    } else {
      showError(message);
    }
  };

  const getRebates = async (currentPage = page, currentPageSize = pageSize) => {
    setLoading(true);
    try {
      const res = await API.get(
        `/api/user/aff/rebates?p=${currentPage}&page_size=${currentPageSize}`,
      );
      const { success, message, data } = res.data;
      if (success) {
        setRebates(data.items || []);
        setTotal(data.total || 0);
      } else {
        showError(message);
      }
    } catch (e) {
      showError(t('加载失败'));
    } finally {
      setLoading(false);
    }
  };

  const transfer = async () => {
    if (transferAmount < getQuotaPerUnit()) {
      showError(t('划转金额最低为') + ' ' + renderQuota(getQuotaPerUnit()));
      return;
    }
    const res = await API.post('/api/user/aff_transfer', {
      quota: transferAmount,
    });
    const { success, message } = res.data;
    if (success) {
      showSuccess(message);
      setOpenTransfer(false);
      getUserQuota().then();
    } else {
      showError(message);
    }
  };

  const handleAffLinkClick = async () => {
    await copy(affLink);
    showSuccess(t('邀请链接已复制到剪切板'));
  };

  useEffect(() => {
    getUserQuota().then();
    setTransferAmount(getQuotaPerUnit());
  }, []);

  useEffect(() => {
    if (affFetchedRef.current) return;
    affFetchedRef.current = true;
    getAffLink().then();
  }, []);

  useEffect(() => {
    getRebates(page, pageSize).then();
  }, [page, pageSize]);

  const columns = useMemo(
    () => [
      {
        title: t('类型'),
        dataIndex: 'source_type',
        key: 'source_type',
        render: (type) => (
          <Tag
            color={type === 'subscription' ? 'purple' : 'green'}
            shape='circle'
          >
            {type === 'subscription' ? t('订阅') : t('充值')}
          </Tag>
        ),
      },
      {
        title: t('来源订单'),
        dataIndex: 'source_id',
        key: 'source_id',
        render: (sourceId) => <Text copyable>{sourceId}</Text>,
      },
      {
        title: t('支付方式'),
        dataIndex: 'payment_method',
        key: 'payment_method',
        render: (method) => <Text>{method || '-'}</Text>,
      },
      {
        title: t('返利基准额度'),
        dataIndex: 'base_quota',
        key: 'base_quota',
        render: (quota) => <Text>{renderQuota(quota || 0)}</Text>,
      },
      {
        title: t('返利额度'),
        dataIndex: 'rebate_quota',
        key: 'rebate_quota',
        render: (quota) => <Text strong>{renderQuota(quota || 0)}</Text>,
      },
      {
        title: t('返利比例'),
        dataIndex: 'rebate_ratio',
        key: 'rebate_ratio',
        render: (ratio) => (
          <Text>{((Number(ratio) || 0) * 100).toFixed(2)}%</Text>
        ),
      },
      {
        title: t('入账时间'),
        dataIndex: 'created_at',
        key: 'created_at',
        render: (time) => timestamp2string(time),
      },
    ],
    [t],
  );

  return (
    <div className='w-full max-w-7xl mx-auto relative min-h-screen lg:min-h-0 mt-[60px] px-2'>
      <TransferModal
        t={t}
        openTransfer={openTransfer}
        transfer={transfer}
        handleTransferCancel={() => setOpenTransfer(false)}
        userState={userState}
        renderQuota={renderQuota}
        getQuotaPerUnit={getQuotaPerUnit}
        transferAmount={transferAmount}
        setTransferAmount={setTransferAmount}
      />

      <div className='grid grid-cols-1 gap-6'>
        <InvitationCard
          t={t}
          userState={userState}
          renderQuota={renderQuota}
          setOpenTransfer={setOpenTransfer}
          affLink={affLink}
          handleAffLinkClick={handleAffLinkClick}
          inviteRebateEnabled={inviteRebateEnabled}
          inviteRebatePercent={inviteRebatePercent}
        />

        <Card className='!rounded-2xl shadow-sm border-0' title={t('返利明细')}>
          <Table
            columns={columns}
            dataSource={rebates}
            loading={loading}
            rowKey='id'
            size='small'
            pagination={{
              currentPage: page,
              pageSize,
              total,
              showSizeChanger: true,
              pageSizeOpts: [10, 20, 50, 100],
              onPageChange: setPage,
              onPageSizeChange: (nextPageSize) => {
                setPageSize(nextPageSize);
                setPage(1);
              },
            }}
            empty={
              <Empty
                image={
                  <IllustrationNoResult style={{ width: 150, height: 150 }} />
                }
                darkModeImage={
                  <IllustrationNoResultDark
                    style={{ width: 150, height: 150 }}
                  />
                }
                description={t('暂无返利记录')}
                style={{ padding: 30 }}
              />
            }
          />
        </Card>
      </div>
    </div>
  );
};

export default Aff;
