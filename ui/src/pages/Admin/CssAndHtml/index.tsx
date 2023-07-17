import { FC, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

import type * as Type from '@/common/interface';
import { getPageCustom, putPageCustom } from '@/services';
import { SchemaForm, JSONSchema, initFormData, UISchema } from '@/components';
import { useToast } from '@/hooks';
import { handleFormError } from '@/utils';
import { customizeStore } from '@/stores';

const Index: FC = () => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'admin.css_and_html',
  });
  const Toast = useToast();
  const schema: JSONSchema = {
    title: t('page_title'),
    properties: {
      custom_css: {
        type: 'string',
        title: t('custom_css.label'),
        description: t('custom_css.text'),
      },
      custom_head: {
        type: 'string',
        title: t('head.label'),
        description: t('head.text'),
      },
      custom_header: {
        type: 'string',
        title: t('header.label'),
        description: t('header.text'),
      },
      custom_sidebar: {
        type: 'string',
        title: t('sidebar.label'),
        description: t('sidebar.text'),
      },
      custom_footer: {
        type: 'string',
        title: t('footer.label'),
        description: t('footer.text'),
      },
    },
  };
  const uiSchema: UISchema = {
    custom_css: {
      'ui:widget': 'textarea',
      'ui:options': {
        rows: 10,
        className: ['small', 'font-monospace'],
      },
    },
    custom_head: {
      'ui:widget': 'textarea',
      'ui:options': {
        rows: 10,
        className: ['small', 'font-monospace'],
      },
    },
    custom_header: {
      'ui:widget': 'textarea',
      'ui:options': {
        rows: 10,
        className: ['small', 'font-monospace'],
      },
    },
    custom_sidebar: {
      'ui:widget': 'textarea',
      'ui:options': {
        rows: 10,
        className: ['small', 'font-monospace'],
      },
    },
    custom_footer: {
      'ui:widget': 'textarea',
      'ui:options': {
        rows: 10,
        className: ['small', 'font-monospace'],
      },
    },
  };
  const [formData, setFormData] = useState(initFormData(schema));
  const onSubmit = (evt) => {
    evt.preventDefault();
    evt.stopPropagation();

    const reqParams: Type.AdminSettingsCustom = {
      custom_css: formData.custom_css.value,
      custom_head: formData.custom_head.value,
      custom_header: formData.custom_header.value,
      custom_sidebar: formData.custom_sidebar.value,
      custom_footer: formData.custom_footer.value,
    };

    putPageCustom(reqParams)
      .then(() => {
        Toast.onShow({
          msg: t('update', { keyPrefix: 'toast' }),
          variant: 'success',
        });
        customizeStore.getState().update(reqParams);
      })
      .catch((err) => {
        if (err.isError) {
          const data = handleFormError(err, formData);
          setFormData({ ...data });
        }
      });
  };

  useEffect(() => {
    getPageCustom().then((setting) => {
      if (setting) {
        const formMeta = { ...formData };
        formMeta.custom_css.value = setting.custom_css;
        formMeta.custom_head.value = setting.custom_head;
        formMeta.custom_header.value = setting.custom_header;
        formMeta.custom_sidebar.value = setting.custom_sidebar;
        formMeta.custom_footer.value = setting.custom_footer;
        setFormData(formMeta);
      }
    });
  }, []);

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

export default Index;
