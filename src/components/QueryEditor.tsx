import React, { ChangeEvent } from 'react';
import { InlineField, Input, Select, Stack } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from '../datasource';
import { MyDataSourceOptions, MyQuery } from '../types';
import { getTemplateSrv } from '@grafana/runtime';

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export function QueryEditor({ query, onChange, onRunQuery }: Props) {
  const templateSrv = getTemplateSrv();

  const variableOptions = templateSrv.getVariables().map((v) => ({
    label: `$${v.name}`,
    value: `$${v.name}`,
  }));

  const onTypeChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, type: event.target.value });
  };

  return (
    <Stack gap={0}>
      <InlineField label="Detalhamento">
        <Select
          options={variableOptions}
          value={query.det}
          defaultValue={"Dias"}
          onChange={(v) => onChange({ ...query, det: String(v.value) })}
        />
      </InlineField>
      <InlineField label="Tipo">
        <Input
          id="query-editor-type"
          onChange={onTypeChange}
          value={query.type}
          required
          placeholder='database | compute | compute-all'
        />
      </InlineField>
    </Stack>
  );
}
