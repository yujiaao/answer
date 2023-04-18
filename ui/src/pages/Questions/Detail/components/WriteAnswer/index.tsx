import { memo, useState, FC, useEffect } from 'react';
import { Form, Button, Alert } from 'react-bootstrap';
import { useTranslation, Trans } from 'react-i18next';

import { marked } from 'marked';
import classNames from 'classnames';

import { usePromptWithUnload } from '@/hooks';
import { Editor, Modal, TextArea } from '@/components';
import { FormDataType } from '@/common/interface';
import { postAnswer } from '@/services';
import { guard, handleFormError, SaveDraft, storageExpires } from '@/utils';
import { DRAFT_ANSWER_STORAGE_KEY } from '@/common/constants';

interface Props {
  visible?: boolean;
  data: {
    /** question  id */
    qid: string;
    answered?: boolean;
    loggedUserRank: number;
  };
  callback?: (obj) => void;
}

const saveDraft = new SaveDraft({ type: 'answer' });

const Index: FC<Props> = ({ visible = false, data, callback }) => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'question_detail.write_answer',
  });
  const [formData, setFormData] = useState<FormDataType>({
    content: {
      value: '',
      isInvalid: false,
      errorMsg: '',
    },
  });
  const [showEditor, setShowEditor] = useState<boolean>(visible);
  const [focusType, setFocusType] = useState('');
  const [editorFocusState, setEditorFocusState] = useState(false);
  const [hasDraft, setHasDraft] = useState(false);
  const [showTips, setShowTips] = useState(data.loggedUserRank < 100);

  usePromptWithUnload({
    when: Boolean(formData.content.value),
  });

  const removeDraft = () => {
    // immediately remove debounced save
    saveDraft.save.cancel();
    saveDraft.remove();
    setHasDraft(false);
  };

  useEffect(() => {
    const draft = storageExpires.get(DRAFT_ANSWER_STORAGE_KEY);
    if (draft?.questionId === data.qid && draft?.content) {
      setFormData({
        content: {
          value: draft.content,
          isInvalid: false,
          errorMsg: '',
        },
      });
      setShowEditor(true);
      setHasDraft(true);
    }
  }, []);

  useEffect(() => {
    const draft = storageExpires.get(DRAFT_ANSWER_STORAGE_KEY);
    const { content } = formData;

    if (content.value) {
      // save Draft
      saveDraft.save({
        questionId: data?.qid,
        content: content.value,
      });

      setHasDraft(true);
    } else if (draft?.questionId === data.qid && !content.value) {
      removeDraft();
    }
  }, [formData.content.value]);

  const checkValidated = (): boolean => {
    let bol = true;
    const { content } = formData;

    if (!content.value || Array.from(content.value.trim()).length < 6) {
      bol = false;
      formData.content = {
        value: content.value,
        isInvalid: true,
        errorMsg: t('characters'),
      };
    } else {
      formData.content = {
        value: content.value,
        isInvalid: false,
        errorMsg: '',
      };
    }

    setFormData({
      ...formData,
    });
    return bol;
  };

  const resetForm = () => {
    setFormData({
      content: {
        value: '',
        isInvalid: false,
        errorMsg: '',
      },
    });
  };

  const deleteDraft = () => {
    const res = window.confirm(t('discard_confirm', { keyPrefix: 'draft' }));
    if (res) {
      removeDraft();
      resetForm();
    }
  };

  const handleSubmit = () => {
    if (!guard.tryNormalLogged(true)) {
      return;
    }
    if (!checkValidated()) {
      return;
    }
    postAnswer({
      question_id: data?.qid,
      content: formData.content.value,
      html: marked.parse(formData.content.value),
    })
      .then((res) => {
        setShowEditor(false);
        setFormData({
          content: {
            value: '',
            isInvalid: false,
            errorMsg: '',
          },
        });
        removeDraft();
        callback?.(res.info);
      })
      .catch((ex) => {
        if (ex.isError) {
          const stateData = handleFormError(ex, formData);
          setFormData({ ...stateData });
        }
      });
  };

  const clickBtn = () => {
    if (!guard.tryNormalLogged(true)) {
      return;
    }
    if (data?.answered && !showEditor) {
      Modal.confirm({
        title: t('confirm_title'),
        content: t('confirm_info'),
        confirmText: t('continue'),
        onConfirm: () => {
          setShowEditor(true);
        },
      });
      return;
    }

    if (!showEditor) {
      setShowEditor(true);
      return;
    }

    handleSubmit();
  };
  const handleFocusForTextArea = (evt) => {
    if (!guard.tryNormalLogged(true)) {
      evt.currentTarget.blur();
      return;
    }
    setFocusType('answer');
    setShowEditor(true);
    setEditorFocusState(true);
  };
  return (
    <Form noValidate className="mt-4">
      {(!data.answered || showEditor) && (
        <Form.Group className="mb-3">
          <Form.Label>
            <h5>{t('title')}</h5>
          </Form.Label>
          <Form.Control
            isInvalid={formData.content.isInvalid}
            className="d-none"
          />
          {!showEditor && !data.answered && (
            <div className="d-flex">
              <TextArea
                className="w-100"
                rows={8}
                autoFocus={false}
                onFocus={handleFocusForTextArea}
              />
            </div>
          )}
          {showEditor && (
            <>
              <Editor
                className={classNames(
                  'form-control p-0',
                  focusType === 'answer' && 'focus',
                )}
                value={formData.content.value}
                autoFocus={editorFocusState}
                onChange={(val) => {
                  setFormData({
                    content: {
                      value: val,
                      isInvalid: false,
                      errorMsg: '',
                    },
                  });
                }}
                onFocus={() => {
                  setFocusType('answer');
                }}
                onBlur={() => {
                  setFocusType('');
                }}
              />

              <Alert
                variant="warning"
                show={data.loggedUserRank < 100 && showTips}
                onClose={() => setShowTips(false)}
                dismissible
                className="mt-3">
                <p>{t('tips.header_1')}</p>
                <ul>
                  <li>
                    <Trans
                      i18nKey="question_detail.write_answer.tips.li1_1"
                      components={{ strong: <strong /> }}
                    />
                  </li>
                  <li>{t('tips.li1_2')}</li>
                </ul>
                <p>
                  <Trans
                    i18nKey="question_detail.write_answer.tips.header_2"
                    components={{ strong: <strong /> }}
                  />
                </p>
                <ul className="mb-0">
                  <li>{t('tips.li2_1')}</li>
                </ul>
              </Alert>
            </>
          )}

          <Form.Control.Feedback type="invalid">
            {formData.content.errorMsg}
          </Form.Control.Feedback>
        </Form.Group>
      )}

      {data.answered && !showEditor ? (
        <Button onClick={clickBtn}>{t('add_another_answer')}</Button>
      ) : (
        <Button onClick={clickBtn}>{t('btn_name')}</Button>
      )}
      {hasDraft && (
        <Button variant="link" className="ms-2" onClick={deleteDraft}>
          {t('discard_draft', { keyPrefix: 'btns' })}
        </Button>
      )}
    </Form>
  );
};

export default memo(Index);
