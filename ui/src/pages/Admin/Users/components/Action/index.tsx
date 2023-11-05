/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import { Dropdown } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import { Modal, Icon } from '@/components';
import {
  useChangeUserRoleModal,
  useChangePasswordModal,
  useActivationEmailModal,
  useToast,
} from '@/hooks';
import { updateUserPassword, changeUserStatus } from '@/services';

interface Props {
  showActionPassword?: boolean;
  showActionStatus?: boolean;
  showActionRole?: boolean;
  currentUser;
  refreshUsers: () => void;
  showDeleteModal: (val) => void;
  userData;
}

const UserOperation = ({
  showActionPassword,
  showActionStatus,
  showActionRole,
  currentUser,
  refreshUsers,
  showDeleteModal,
  userData,
}: Props) => {
  const { t } = useTranslation('translation', { keyPrefix: 'admin.users' });
  const Toast = useToast();

  const changeUserRoleModal = useChangeUserRoleModal({
    callback: refreshUsers,
  });
  const changePasswordModal = useChangePasswordModal({
    onConfirm: (rd) => {
      return new Promise((resolve, reject) => {
        updateUserPassword(rd)
          .then(() => {
            Toast.onShow({
              msg: t('update_password', { keyPrefix: 'toast' }),
              variant: 'success',
            });
            resolve(true);
          })
          .catch((e) => {
            reject(e);
          });
      });
    },
  });

  const activationEmailModal = useActivationEmailModal();

  const postUserStatus = (statusType) => {
    changeUserStatus({
      user_id: userData.user_id,
      status: statusType,
    }).then(() => {
      refreshUsers?.();
      // onClose();
    });
  };

  const handleAction = (type) => {
    const { user_id, role_id, username } = userData;
    if (username === currentUser.username) {
      Toast.onShow({
        msg: t('forbidden_operate_self', { keyPrefix: 'toast' }),
        variant: 'warning',
      });
      return;
    }

    if (type === 'role') {
      changeUserRoleModal.onShow({
        id: user_id,
        role_id,
      });
    }

    if (type === 'password') {
      changePasswordModal.onShow(user_id);
    }

    if (type === 'activation') {
      activationEmailModal.onShow(user_id);
    }

    if (type === 'deactivate') {
      // cons
      Modal.confirm({
        title: t('deactivate_user.title'),
        content: t('deactivate_user.content'),
        cancelBtnVariant: 'link',
        confirmBtnVariant: 'danger',
        cancelText: t('cancel', { keyPrefix: 'btns' }),
        confirmText: t('deactivate', { keyPrefix: 'btns' }),
        onConfirm: () => {
          // active -> inactive
          postUserStatus('inactive');
        },
      });
    }

    if (type === 'suspend') {
      // cons
      Modal.confirm({
        title: t('suspend_user.title'),
        content: t('suspend_user.content'),
        cancelBtnVariant: 'link',
        cancelText: t('cancel', { keyPrefix: 'btns' }),
        confirmBtnVariant: 'danger',
        confirmText: t('suspend', { keyPrefix: 'btns' }),
        onConfirm: () => {
          // active -> suspended
          postUserStatus('suspended');
        },
      });
    }

    if (type === 'active' || type === 'unsuspend') {
      // to normal
      postUserStatus('normal');
    }

    if (type === 'delete') {
      showDeleteModal({
        show: true,
        userId: userData.user_id,
      });
    }
  };

  return (
    <td className="text-end">
      <Dropdown>
        <Dropdown.Toggle variant="link" className="no-toggle p-0">
          <Icon name="three-dots-vertical" title={t('action')} />
        </Dropdown.Toggle>
        <Dropdown.Menu>
          {showActionPassword ? (
            <Dropdown.Item onClick={() => handleAction('password')}>
              {t('set_new_password')}
            </Dropdown.Item>
          ) : null}
          {showActionRole ? (
            <Dropdown.Item onClick={() => handleAction('role')}>
              {t('change_role')}
            </Dropdown.Item>
          ) : null}
          {userData.status === 'inactive' ? (
            <Dropdown.Item onClick={() => handleAction('activation')}>
              {t('btn_name', { keyPrefix: 'inactive' })}
            </Dropdown.Item>
          ) : null}
          {showActionStatus && userData.status !== 'deleted' ? (
            <>
              <Dropdown.Divider />
              {userData.status === 'inactive' && (
                <Dropdown.Item onClick={() => handleAction('active')}>
                  {t('active', { keyPrefix: 'btns' })}
                </Dropdown.Item>
              )}
              {userData.status === 'normal' && (
                <Dropdown.Item onClick={() => handleAction('deactivate')}>
                  {t('deactivate', { keyPrefix: 'btns' })}
                </Dropdown.Item>
              )}
              {userData.status === 'normal' && (
                <Dropdown.Item onClick={() => handleAction('suspend')}>
                  {t('suspend', { keyPrefix: 'btns' })}
                </Dropdown.Item>
              )}
              {userData.status === 'suspended' && (
                <Dropdown.Item onClick={() => handleAction('unsuspend')}>
                  {t('unsuspend', { keyPrefix: 'btns' })}
                </Dropdown.Item>
              )}
              <Dropdown.Item onClick={() => handleAction('delete')}>
                {t('delete', { keyPrefix: 'btns' })}
              </Dropdown.Item>
            </>
          ) : null}
        </Dropdown.Menu>
      </Dropdown>
    </td>
  );
};

export default UserOperation;
