import React, { FC, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

import { SchemaForm, JSONSchema, initFormData, UISchema } from '@/components';
import type * as Type from '@/common/interface';
import { useToast } from '@/hooks';
import { siteInfoStore } from '@/stores';
import { useGeneralSetting, updateGeneralSetting } from '@/services';
import Pattern from '@/common/pattern';
import { handleFormError } from '@/utils';

const General: FC = () => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'admin.general',
  });
  const Toast = useToast();
  const updateSiteInfo = siteInfoStore((state) => state.update);

  const { data: setting } = useGeneralSetting();
  const schema: JSONSchema = {
    title: t('page_title'),
    required: ['name', 'site_url', 'contact_email'],
    properties: {
      name: {
        type: 'string',
        title: t('name.label'),
        description: t('name.text'),
      },
      site_url: {
        type: 'string',
        title: t('site_url.label'),
        description: t('site_url.text'),
      },
      short_description: {
        type: 'string',
        title: `${t('short_desc.label')} ${t('optional', {
          keyPrefix: 'form',
        })}`,
        description: t('short_desc.text'),
      },
      description: {
        type: 'string',
        title: `${t('desc.label')} ${t('optional', {
          keyPrefix: 'form',
        })}`,
        description: t('desc.text'),
      },
      contact_email: {
        type: 'string',
        title: t('contact_email.label'),
        description: t('contact_email.text'),
      },
    },
  };
  const uiSchema: UISchema = {
    site_url: {
      'ui:options': {
        inputType: 'url',
        validator: (value) => {
          let url: URL | undefined;
          try {
            url = new URL(value);
          } catch (ex) {
            return t('site_url.validate');
          }
          if (
            !url ||
            !/^https?:$/.test(url.protocol) ||
            url.pathname !== '/' ||
            url.search !== '' ||
            url.hash !== ''
          ) {
            return t('site_url.validate');
          }

          return true;
        },
      },
    },
    contact_email: {
      'ui:options': {
        inputType: 'email',
        validator: (value) => {
          if (!Pattern.email.test(value)) {
            return t('contact_email.validate');
          }
          return true;
        },
      },
    },
  };
  const [formData, setFormData] = useState<Type.FormDataType>(
    initFormData(schema),
  );

  const onSubmit = (evt) => {
    evt.preventDefault();
    evt.stopPropagation();
    const reqParams: Type.AdminSettingsGeneral = {
      name: formData.name.value,
      description: formData.description.value,
      short_description: formData.short_description.value,
      site_url: formData.site_url.value,
      contact_email: formData.contact_email.value,
    };

    updateGeneralSetting(reqParams)
      .then((res) => {
        Toast.onShow({
          msg: t('update', { keyPrefix: 'toast' }),
          variant: 'success',
        });
        if (res.name) {
          formData.name.value = res.name;
          formData.description.value = res.description;
          formData.short_description.value = res.short_description;
          formData.site_url.value = res.site_url;
          formData.contact_email.value = res.contact_email;
        }

        setFormData({ ...formData });
        updateSiteInfo(res);
      })
      .catch((err) => {
        if (err.isError) {
          const data = handleFormError(err, formData);
          setFormData({ ...data });
        }
      });
  };

  useEffect(() => {
    if (!setting) {
      return;
    }
    const formMeta: Type.FormDataType = {};
    Object.keys(formData).forEach((k) => {
      formMeta[k] = { ...formData[k], value: setting[k] };
    });
    setFormData({ ...formData, ...formMeta });
  }, [setting]);

  const handleOnChange = (data) => {
    setFormData(data);
  };

  return (
    <>
      <h3 className="mb-4">{t('page_title')}</h3>
      <SchemaForm
        schema={schema}
        formData={formData}
        onSubmit={onSubmit}
        uiSchema={uiSchema}
        onChange={handleOnChange}
      />
    </>
  );
};

export default General;
