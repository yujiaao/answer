import { memo, FC, useEffect, useRef } from 'react';
import { Button, Alert, Badge } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';
import { Link, useSearchParams } from 'react-router-dom';

import {
  Actions,
  Operate,
  UserCard,
  Icon,
  Comment,
  FormatTime,
  htmlRender,
} from '@/components';
import { scrollToElementTop, bgFadeOut } from '@/utils';
import { AnswerItem } from '@/common/interface';
import { acceptanceAnswer } from '@/services';

interface Props {
  data: AnswerItem;
  /** router answer id */
  aid?: string;
  canAccept: boolean;
  questionTitle: string;
  slugTitle: string;
  isLogged: boolean;
  callback: (type: string) => void;
}
const Index: FC<Props> = ({
  aid,
  data,
  isLogged,
  questionTitle = '',
  slugTitle,
  callback,
  canAccept = false,
}) => {
  const { t } = useTranslation('translation', {
    keyPrefix: 'question_detail',
  });
  const [searchParams] = useSearchParams();
  const answerRef = useRef<HTMLDivElement>(null);
  const acceptAnswer = () => {
    acceptanceAnswer({
      question_id: data.question_id,
      answer_id: data.accepted === 2 ? '0' : data.id,
    }).then(() => {
      callback?.('');
    });
  };

  useEffect(() => {
    if (!answerRef?.current) {
      return;
    }

    htmlRender(answerRef.current.querySelector('.fmt'));

    if (aid === data.id) {
      setTimeout(() => {
        const element = answerRef.current;
        scrollToElementTop(element);
        if (!searchParams.get('commentId')) {
          bgFadeOut(answerRef.current);
        }
      }, 100);
    }
  }, [data.id, answerRef.current]);
  if (!data?.id) {
    return null;
  }
  return (
    <div id={data.id} ref={answerRef} className="answer-item py-4">
      {data.status === 10 && (
        <Alert variant="danger" className="mb-4">
          {t('post_deleted', { keyPrefix: 'messages' })}
        </Alert>
      )}
      {data?.accepted === 2 && (
        <div className="mb-3 lh-1">
          <Badge bg="success" pill>
            <Icon name="check-circle-fill me-1" />
            Best answer
          </Badge>
        </div>
      )}
      <article
        className="fmt text-break text-wrap"
        dangerouslySetInnerHTML={{ __html: data?.html }}
      />
      <div className="d-flex align-items-center mt-4">
        <Actions
          source="answer"
          data={{
            id: data?.id,
            isHate: data?.vote_status === 'vote_down',
            isLike: data?.vote_status === 'vote_up',
            votesCount: data?.vote_count,
            hideCollect: true,
            collected: data?.collected,
            collectCount: 0,
            username: data?.user_info?.username,
          }}
        />

        {canAccept && (
          <Button
            variant={data.accepted === 2 ? 'success' : 'outline-success'}
            className="ms-3"
            onClick={acceptAnswer}>
            <Icon name="check-circle-fill" className="me-2" />
            <span>
              {data.accepted === 2
                ? t('answers.btn_accepted')
                : t('answers.btn_accept')}
            </span>
          </Button>
        )}
      </div>

      <div className="d-block d-md-flex flex-wrap mt-4 mb-3">
        <div className="mb-3 mb-md-0 me-4 flex-grow-1">
          <Operate
            qid={data.question_id}
            aid={data.id}
            memberActions={data?.member_actions}
            type="answer"
            isAccepted={data.accepted === 2}
            title={questionTitle}
            slugTitle={slugTitle}
            callback={callback}
          />
        </div>
        <div className="mb-3 mb-md-0 me-4" style={{ minWidth: '196px' }}>
          {data.update_user_info &&
          data.update_user_info?.username !== data.user_info?.username ? (
            <UserCard
              data={data?.update_user_info}
              time={Number(data.update_time)}
              preFix={t('edit')}
              isLogged={isLogged}
              timelinePath={`/posts/${data.question_id}/${data.id}/timeline`}
            />
          ) : isLogged ? (
            <Link to={`/posts/${data.question_id}/${data.id}/timeline`}>
              <FormatTime
                time={Number(data.update_time)}
                preFix={t('edit')}
                className="link-secondary fs-14"
              />
            </Link>
          ) : (
            <FormatTime
              time={Number(data.update_time)}
              preFix={t('edit')}
              className="text-secondary fs-14"
            />
          )}
        </div>
        <div style={{ minWidth: '196px' }}>
          <UserCard
            data={data?.user_info}
            time={Number(data.create_time)}
            preFix={t('answered')}
            isLogged={isLogged}
            timelinePath={`/posts/${data.question_id}/${data.id}/timeline`}
          />
        </div>
      </div>

      <Comment
        objectId={data.id}
        mode="answer"
        commentId={searchParams.get('commentId')}
      />
    </div>
  );
};

export default memo(Index);
