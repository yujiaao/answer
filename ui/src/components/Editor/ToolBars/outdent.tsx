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

import { memo } from 'react';
import { useTranslation } from 'react-i18next';

import ToolItem from '../toolItem';
import { IEditorContext } from '../types';

let context: IEditorContext;
const Outdent = () => {
  const { t } = useTranslation('translation', { keyPrefix: 'editor' });
  const item = {
    label: 'outdent',
    keyMap: ['Shift-Tab'],
    tip: t('outdent.text'),
  };
  const handleClick = (ctx) => {
    context = ctx;
    const { editor, replaceLines } = context;
    replaceLines((line) => {
      line = line.replace(/^(\s{0,})/, (_1, $1) => {
        return $1.length > 4 ? $1.substring(4) : '';
      });
      return line;
    });
    editor?.focus();
  };

  return <ToolItem {...item} onClick={handleClick} />;
};

export default memo(Outdent);
