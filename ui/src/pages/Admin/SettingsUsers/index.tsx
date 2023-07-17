import { FC, FormEvent, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

import { useToast } from '@/hooks';
import { FormDataType } from '@/common/interface';
import { JSONSchema, SchemaForm, UISchema, initFormData } from '@/components';
import { SYSTEM_AVATAR_OPTIONS } from '@/common/constants';
import {
  getUsersSetting,
  putUsersSetting,
  AdminSettingsUsers,
} from '@/services';
import { handleFormError } from '@/utils';
import * as Type from '@/common/interface';
import { siteInfoStore } from '@/stores';

const Index: FC = () => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'admin.settings_users',
  });
  const Toast = useToast();
  const { updateUsers: updateUsersStore } = siteInfoStore();
  const schema: JSONSchema = {
    title: t('title'),
    properties: {
      default_avatar: {
        type: 'string',
        title: t('avatar.label'),
        description: t('avatar.text'),
        enum: SYSTEM_AVATAR_OPTIONS?.map((v) => v.value),
        enumNames: SYSTEM_AVATAR_OPTIONS?.map((v) => v.label),
        default: 'system',
      },
      gravatar_base_url: {
        type: 'string',
        title: t('gravatar_base_url.label'),
        description: t('gravatar_base_url.text'),
      },
      profile_editable: {
        type: 'string',
        title: t('profile_editable.title'),
      },
      allow_update_display_name: {
        type: 'boolean',
        title: 'allow_update_display_name',
      },
      allow_update_username: {
        type: 'boolean',
        title: 'allow_update_username',
      },
      allow_update_avatar: {
        type: 'boolean',
        title: 'allow_update_avatar',
      },
      allow_update_bio: {
        type: 'boolean',
        title: 'allow_update_bio',
      },
      allow_update_website: {
        type: 'boolean',
        title: 'allow_update_website',
      },
      allow_update_location: {
        type: 'boolean',
        title: 'allow_update_location',
      },
    },
  };

  const [formData, setFormData] = useState<FormDataType>(initFormData(schema));

  const uiSchema: UISchema = {
    default_avatar: {
      'ui:widget': 'select',
    },
    gravatar_base_url: {
      'ui:widget': 'input',
    },
    profile_editable: {
      'ui:widget': 'legend',
    },
    allow_update_display_name: {
      'ui:widget': 'switch',
      'ui:options': {
        label: t('allow_update_display_name.label'),
        simplify: true,
      },
    },
    allow_update_username: {
      'ui:widget': 'switch',
      'ui:options': {
        label: t('allow_update_username.label'),
        simplify: true,
      },
    },
    allow_update_avatar: {
      'ui:widget': 'switch',
      'ui:options': {
        label: t('allow_update_avatar.label'),
        simplify: true,
      },
    },
    allow_update_bio: {
      'ui:widget': 'switch',
      'ui:options': {
        label: t('allow_update_bio.label'),
        simplify: true,
      },
    },
    allow_update_website: {
      'ui:widget': 'switch',
      'ui:options': {
        label: t('allow_update_website.label'),
        simplify: true,
      },
    },
    allow_update_location: {
      'ui:widget': 'switch',
      'ui:options': {
        label: t('allow_update_location.label'),
        fieldClassName: 'mb-3',
        simplify: true,
      },
    },
  };

  const onSubmit = (evt: FormEvent) => {
    evt.preventDefault();
    evt.stopPropagation();
    const reqParams: AdminSettingsUsers = {
      allow_update_avatar: formData.allow_update_avatar.value,
      allow_update_bio: formData.allow_update_bio.value,
      allow_update_display_name: formData.allow_update_display_name.value,
      allow_update_location: formData.allow_update_location.value,
      allow_update_username: formData.allow_update_username.value,
      allow_update_website: formData.allow_update_website.value,
      default_avatar: formData.default_avatar.value,
      gravatar_base_url: formData.gravatar_base_url.value,
    };
    putUsersSetting(reqParams)
      .then(() => {
        updateUsersStore(reqParams);
        Toast.onShow({
          msg: t('update', { keyPrefix: 'toast' }),
          variant: 'success',
        });
      })
      .catch((err) => {
        if (err.isError) {
          const data = handleFormError(err, formData);
          setFormData({ ...data });
        }
      });
  };

  useEffect(() => {
    getUsersSetting().then((resp) => {
      if (!resp) {
        return;
      }
      const formMeta: Type.FormDataType = {};
      Object.keys(formData).forEach((k) => {
        let v = resp[k];
        if (k === 'default_avatar' && !v) {
          v = 'system';
        }
        if (k === 'gravatar_base_url' && !v) {
          v = 'https://www.gravatar.com/avatar/';
        }
        formMeta[k] = { ...formData[k], value: v };
      });
      setFormData({ ...formData, ...formMeta });
    });
  }, []);

  const handleOnChange = (data) => {
    setFormData(data);
  };

  return (
    <>
      <h3 className="mb-4">{t('title')}</h3>
      <SchemaForm
        schema={schema}
        uiSchema={uiSchema}
        formData={formData}
        onSubmit={onSubmit}
        onChange={handleOnChange}
      />
    </>
  );
};

export default Index;
