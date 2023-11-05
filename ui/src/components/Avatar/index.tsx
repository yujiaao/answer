/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import { memo, FC } from 'react';

import classNames from 'classnames';

import DefaultAvatar from '@/assets/images/default-avatar.svg';

interface IProps {
  /** avatar url */
  avatar: string | { type: string; gravatar: string; custom: string };
  /** size 48 96 128 256 */
  size: string;
  searchStr?: string;
  className?: string;
  alt: string;
}

const Index: FC<IProps> = ({
  avatar,
  size,
  className,
  searchStr = '',
  alt,
}) => {
  let url = '';
  if (typeof avatar === 'string') {
    if (avatar.length > 1) {
      url = `${avatar}?${searchStr}${
        avatar?.includes('gravatar') ? '&d=identicon' : ''
      }`;
    }
  } else if (avatar?.type === 'gravatar' && avatar.gravatar) {
    url = `${avatar.gravatar}?${searchStr}&d=identicon`;
  } else if (avatar?.type === 'custom' && avatar.custom) {
    url = `${avatar.custom}?${searchStr}`;
  }

  const roundedCls =
    className && className.indexOf('rounded') !== -1 ? '' : 'rounded';

  return (
    <>
      {/* eslint-disable jsx-a11y/no-noninteractive-element-to-interactive-role,jsx-a11y/control-has-associated-label */}
      <img
        src={url || DefaultAvatar}
        width={size}
        height={size}
        className={classNames(roundedCls, className)}
        alt={alt}
      />
    </>
  );
};

export default memo(Index);
