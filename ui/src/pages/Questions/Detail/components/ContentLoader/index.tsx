import { FC, memo } from 'react';

const Index: FC = () => {
  return (
    <div className="placeholder-glow">
      <div className="placeholder w-100 h1 mb-3" style={{ height: '34px' }} />

      <div className="placeholder w-75 mb-3" style={{ height: '21px' }} />

      <div
        className="placeholder w-50 d-block align-top mb-4"
        style={{ height: '24px' }}
      />

      <div style={{ marginBottom: '2rem' }}>
        <p>
          <span
            className="placeholder w-100 d-block align-top mb-1"
            style={{ height: '24px' }}
          />
          <span
            className="placeholder w-100 d-block align-top mb-1"
            style={{ height: '24px' }}
          />
          <span
            className="placeholder w-100 d-block align-top mb-1"
            style={{ height: '24px' }}
          />

          <span
            className="placeholder w-100 d-block align-top mb-1"
            style={{ height: '24px' }}
          />
          <span
            className="placeholder w-75 d-block align-top"
            style={{ height: '24px' }}
          />
        </p>

        <p>
          <span
            className="placeholder w-100 d-block align-top mb-1"
            style={{ height: '24px' }}
          />
          <span
            className="placeholder w-100 d-block align-top mb-1"
            style={{ height: '24px' }}
          />
          <span
            className="placeholder w-100 d-block align-top mb-1"
            style={{ height: '24px' }}
          />
          <span
            className="placeholder w-100 d-block align-top mb-1"
            style={{ height: '24px' }}
          />
          <span
            className="placeholder w-50 d-block align-top"
            style={{ height: '24px' }}
          />
        </p>
      </div>

      <div className="d-flex">
        <div
          className="placeholder align-top me-3 rounded"
          style={{ height: '38px', width: '120px' }}
        />
        <div
          className="placeholder align-top rounded"
          style={{ height: '38px', width: '68px' }}
        />
      </div>

      <div className="d-block d-md-flex flex-wrap mt-4 mb-3">
        <div
          className="placeholder mb-3 mb-md-0 me-4"
          style={{ height: '21px', width: '40%' }}
        />
        <div
          style={{ minWidth: '196px', height: '24px' }}
          className="placeholder mb-3 me-4 mb-md-0 d-block d-md-none"
        />

        <div
          style={{ minWidth: '196px', height: '24px' }}
          className="placeholder d-block d-md-none"
        />

        <div
          style={{ minWidth: '196px', height: '40px' }}
          className="placeholder mb-3 me-4 mb-md-0 d-none d-md-block"
        />

        <div
          style={{ minWidth: '196px', height: '40px' }}
          className="placeholder d-none d-md-block"
        />
      </div>

      {[0, 1, 2].map((item, i) => (
        <div
          className={`border-bottom py-2 ${i === 0 ? 'border-top' : ''}`}
          key={item}>
          <div className="placeholder w-100 mb-1" style={{ height: '17px' }} />
          <div className="placeholder w-50" style={{ height: '17px' }} />
        </div>
      ))}

      <div className="d-flex mt-2 mb-4">
        <div
          className="placeholder align-top me-4"
          style={{ height: '21px', width: '140px' }}
        />
        <div
          className="placeholder align-top"
          style={{ height: '21px', width: '140px' }}
        />
      </div>
    </div>
  );
};

export default memo(Index);
