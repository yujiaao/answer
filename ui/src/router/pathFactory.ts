import urlcat from 'urlcat';

import { seoSettingStore } from '@/stores';

const tagLanding = (slugName: string) => {
  if (!slugName) {
    return '/tags';
  }
  slugName = slugName.toLowerCase();
  return urlcat('/tags/:slugName', { slugName });
};
const tagInfo = (slugName: string) => {
  if (!slugName) {
    return '/tags';
  }
  slugName = slugName.toLowerCase();
  return urlcat('/tags/:slugName/info', { slugName });
};
const tagEdit = (tagId: string) => {
  return urlcat('/tags/:tagId/edit', { tagId });
};
const questionLanding = (questionId: string, slugTitle: string = '') => {
  const { seo } = seoSettingStore.getState();
  if (seo.permalink === 1 && slugTitle) {
    return urlcat('/questions/:questionId/:slugPermalink', {
      questionId,
      slugPermalink: slugTitle,
    });
  }

  return urlcat('/questions/:questionId', { questionId });
};
const answerLanding = (params: {
  questionId: string;
  slugTitle?: string;
  answerId: string;
}) => {
  const questionLandingUrl = questionLanding(
    params.questionId,
    params.slugTitle,
  );
  return urlcat(`${questionLandingUrl}/:answerId`, {
    answerId: params.answerId,
  });
};

export const pathFactory = {
  tagLanding,
  tagInfo,
  tagEdit,
  questionLanding,
  answerLanding,
};
