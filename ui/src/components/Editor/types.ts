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

import { EditorView, Command } from '@codemirror/view';

export interface Position {
  ch: number;
  line: number;
  sticky?: string | undefined;
}
export interface ExtendEditor {
  addKeyMap: (keyMap: Record<string, Command>) => void;
  on: (
    event:
      | 'change'
      | 'focus'
      | 'blur'
      | 'dragenter'
      | 'dragover'
      | 'drop'
      | 'paste',
    callback: (e?) => void,
  ) => void;
  getValue: () => string;
  setValue: (value: string) => void;
  off: (
    event:
      | 'change'
      | 'focus'
      | 'blur'
      | 'dragenter'
      | 'dragover'
      | 'drop'
      | 'paste',
    callback: (e?) => void,
  ) => void;
  getSelection: () => string;
  replaceSelection: (value: string) => void;
  focus: () => void;
  wrapText: (before: string, after?: string, defaultText?: string) => void;
  replaceLines: (
    replace: Parameters<Array<string>['map']>[0],
    symbolLen?: number,
  ) => void;
  appendBlock: (content: string) => void;
  getCursor: () => Position;
  replaceRange: (value: string, from: Position, to: Position) => void;
  setSelection: (anchor: Position, head?: Position) => void;
  setReadOnly: (readOnly: boolean) => void;
}

export type Editor = EditorView & ExtendEditor;
export interface CodeMirrorEditor extends Editor {
  display: any;

  moduleType;
}

export interface IEditorContext {
  editor: Editor;
  wrapText?;
  replaceLines?;
  appendBlock?;
}
