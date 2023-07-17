import React, { FC } from 'react';
import { Card, Col, Carousel } from 'react-bootstrap';
import { useTranslation } from 'react-i18next';

import { userCenterStore } from '@/stores';

const data = [
  {
    id: 1,
    url: require('@/assets/images/carousel-wecom-1.jpg'),
  },
  {
    id: 2,
    url: require('@/assets/images/carousel-wecom-2.jpg'),
  },
  {
    id: 3,
    url: require('@/assets/images/carousel-wecom-3.jpg'),
  },
  {
    id: 4,
    url: require('@/assets/images/carousel-wecom-4.jpg'),
  },
  {
    id: 5,
    url: require('@/assets/images/carousel-wecom-5.jpg'),
  },
];

const Index: FC = () => {
  const { t } = useTranslation('translation', { keyPrefix: 'user_center' });
  const ucAgent = userCenterStore().agent;
  return (
    <Col lg={4} className="mx-auto mt-3 py-5">
      <Card>
        <Card.Body>
          <h3 className="text-center pt-3 mb-3">
            {ucAgent?.agent_info?.display_name} {t('login')}
          </h3>
          <p className="text-danger text-center">
            {t('login_failed_email_tip')}
          </p>

          <Carousel controls={false}>
            {data.map((item) => (
              <Carousel.Item key={item.id}>
                <img
                  className="d-block w-100"
                  src={item.url}
                  alt="First slide"
                />
              </Carousel.Item>
            ))}
          </Carousel>
        </Card.Body>
      </Card>
    </Col>
  );
};

export default Index;
