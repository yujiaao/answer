import useSWR from 'swr';

import request from '@/utils/request';
import type * as Type from '@/common/interface';

export const useGeneralSetting = () => {
  const apiUrl = `/answer/admin/api/siteinfo/general`;
  const { data, error } = useSWR<Type.AdminSettingsGeneral, Error>(
    [apiUrl],
    request.instance.get,
  );

  return {
    data,
    isLoading: !data && !error,
    error,
  };
};

export const updateGeneralSetting = (params: Type.AdminSettingsGeneral) => {
  const apiUrl = `/answer/admin/api/siteinfo/general`;
  return request.put(apiUrl, params);
};

export const useInterfaceSetting = () => {
  const apiUrl = `/answer/admin/api/siteinfo/interface`;
  const { data, error } = useSWR<Type.AdminSettingsInterface, Error>(
    [apiUrl],
    request.instance.get,
  );
  return {
    data,
    isLoading: !data && !error,
    error,
  };
};

export const updateInterfaceSetting = (params: Type.AdminSettingsInterface) => {
  const apiUrl = `/answer/admin/api/siteinfo/interface`;
  return request.put(apiUrl, params);
};

export const useSmtpSetting = () => {
  const apiUrl = `/answer/admin/api/setting/smtp`;
  const { data, error } = useSWR<Type.AdminSettingsSmtp, Error>(
    [apiUrl],
    request.instance.get,
  );
  return {
    data,
    isLoading: !data && !error,
    error,
  };
};

export const updateSmtpSetting = (params: Type.AdminSettingsSmtp) => {
  const apiUrl = `/answer/admin/api/setting/smtp`;
  return request.put(apiUrl, params);
};

export const getAdminLanguageOptions = () => {
  const apiUrl = `/answer/admin/api/language/options`;
  return request.get<Type.LangsType[]>(apiUrl);
};

export const getBrandSetting = () => {
  return request.get('/answer/admin/api/siteinfo/branding');
};

export const brandSetting = (params: Type.AdminSettingBranding) => {
  return request.put('/answer/admin/api/siteinfo/branding', params);
};

export const getRequireAndReservedTag = () => {
  return request.get('/answer/admin/api/siteinfo/write');
};

export const postRequireAndReservedTag = (params) => {
  return request.put('/answer/admin/api/siteinfo/write', params);
};

export const getLegalSetting = () => {
  return request.get<Type.AdminSettingsLegal>(
    '/answer/admin/api/siteinfo/legal',
  );
};

export const putLegalSetting = (params: Type.AdminSettingsLegal) => {
  return request.put('/answer/admin/api/siteinfo/legal', params);
};

export const getSeoSetting = () => {
  return request.get<Type.AdminSettingsSeo>('/answer/admin/api/siteinfo/seo');
};

export const putSeoSetting = (params: Type.AdminSettingsSeo) => {
  return request.put('/answer/admin/api/siteinfo/seo', params);
};

export const getThemeSetting = () => {
  return request.get<Type.AdminSettingsTheme>(
    '/answer/admin/api/siteinfo/theme',
  );
};

export const putThemeSetting = (params: Type.AdminSettingsTheme) => {
  return request.put('/answer/admin/api/siteinfo/theme', params);
};

export const getPageCustom = () => {
  return request.get<Type.AdminSettingsCustom>(
    '/answer/admin/api/siteinfo/custom-css-html',
  );
};

export const putPageCustom = (params: Type.AdminSettingsCustom) => {
  return request.put('/answer/admin/api/siteinfo/custom-css-html', params);
};

export const getLoginSetting = () => {
  return request.get<Type.AdminSettingsLogin>(
    '/answer/admin/api/siteinfo/login',
  );
};

export const putLoginSetting = (params: Type.AdminSettingsLogin) => {
  return request.put('/answer/admin/api/siteinfo/login', params);
};
