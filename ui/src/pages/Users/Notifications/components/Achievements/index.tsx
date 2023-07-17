import { ListGroup } from 'react-bootstrap';
import { Link } from 'react-router-dom';

import classNames from 'classnames';
import { isEmpty } from 'lodash';

import { Empty } from '@/components';

import './index.scss';

const Achievements = ({ data, handleReadNotification }) => {
  if (!data) {
    return null;
  }
  if (isEmpty(data)) {
    return <Empty />;
  }
  return (
    <ListGroup className="achievement-wrap rounded-0">
      {data.map((item) => {
        const { comment, question, answer } =
          item?.object_info?.object_map || {};
        let url = '';
        switch (item.object_info.object_type) {
          case 'question':
            url = `/questions/${item.object_info.object_id}`;
            break;
          case 'answer':
            url = `/questions/${question}/${item.object_info.object_id}`;
            break;
          case 'comment':
            url = `/questions/${question}/${answer}?commentId=${comment}`;
            break;
          default:
            url = '';
        }
        return (
          <ListGroup.Item
            key={item.id}
            className={classNames(
              'd-flex border-start-0 border-end-0 py-3',
              !item.is_read && 'warning',
            )}>
            {item.rank > 0 && (
              <div className="text-success num text-end">{`+${item.rank}`}</div>
            )}
            {item.rank === 0 && <div className="num text-end">{item.rank}</div>}
            {item.rank < 0 && (
              <div className="text-danger num text-end">{`${item.rank}`}</div>
            )}
            <div className="d-flex flex-column ms-3 flex-fill">
              <Link to={url} onClick={() => handleReadNotification(item.id)}>
                {item.object_info.title}
              </Link>
              <span className="text-secondary small">
                {item.object_info.object_type}
              </span>
            </div>
          </ListGroup.Item>
        );
      })}
    </ListGroup>
  );
};

export default Achievements;
