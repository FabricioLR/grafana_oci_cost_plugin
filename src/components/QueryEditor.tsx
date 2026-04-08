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

  const variableOptions2 = ["all", "database", "compute", "compute-all"].map((v) => ({
    label: `${v}`,
    value: `${v}`,
  }));

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
        <Select
          options={variableOptions2}
          value={query.type}
          defaultValue={"all"}
          onChange={(v) => onChange({ ...query, type: String(v.value) })}
        />
      </InlineField>
    </Stack>
  );
}
