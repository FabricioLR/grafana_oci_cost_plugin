import React, { useEffect, useState } from 'react';
import { InlineField, Select, Stack } from '@grafana/ui';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { DataSource } from '../datasource';
import { MyDataSourceOptions, MyQuery } from '../types';
import { getBackendSrv, getTemplateSrv } from '@grafana/runtime';

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export function QueryEditor({ query, onChange, datasource }: Props) {
  const templateSrv = getTemplateSrv();

  const variableOptions = templateSrv.getVariables().map((v) => ({
    label: `$${v.name}`,
    value: `$${v.name}`,
  }));

  variableOptions.push({label: "Dias", value: "Dias"})
  variableOptions.push({label: "Meses", value: "Meses"})

  const variableOptions3 = ["All"].map((v) => ({
    label: `${v}`,
    value: `${v}`,
  }));

  const variableOptions2 = ["Object Storage", "Key Management", "MySQL", "Virtual Cloud Network", "Database", "Block Storage", "Compute", "Database Management", "Telemetry", "All"].map((v) => ({
    label: `${v}`,
    value: `${v}`,
  }));

  const [namespaceOptions, setNamespaceOptions] = useState<Array<SelectableValue<string>>>([]);
  const [tagOptions, setTagOptions] = useState<Array<SelectableValue<string>>>([]);

  useEffect(() => {
    getBackendSrv().get(`/api/datasources/${datasource.id}/resources/namespaces`).then(res => setNamespaceOptions(res.map((n: string) => ({ label: n, value: n }))));
  }, [datasource.id]);

  useEffect(() => {
    if (!query.namespace) {
      setTagOptions([]);
      return;
    }

    getBackendSrv().get(`/api/datasources/${datasource.id}/resources/tags?namespace=${query.namespace}`)
    .then(res => {
      setTagOptions(res.map((t: string) => ({ label: t, value: t })));
    })
    .catch(() => setTagOptions([]));
  }, [datasource.id, query.namespace]);

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
      <InlineField label="Service">
        <Select
          options={variableOptions2}
          value={query.service}
          onChange={(v) => onChange({ ...query, service: String(v.value) })}
          defaultValue={"All"}
          allowCustomValue
        />
      </InlineField>
      <InlineField label="Namespace">
        <Select
          options={namespaceOptions}
          value={query.namespace}
          onChange={(v) => onChange({ ...query, namespace: String(v.value) })}
          defaultValue={"All"}
          placeholder={"Selecione o namespace"}
          allowCustomValue
        />
      </InlineField>
      <InlineField label="Tag">
        <Select
          options={tagOptions}
          value={query.tag_key}
          onChange={(v) => onChange({ ...query, tag_key: String(v.value) })}
          defaultValue={"All"}
          placeholder={"Selecione a tag"}
          allowCustomValue
        />
      </InlineField>
      <InlineField label="Value">
        <Select
          options={variableOptions3}
          value={query.tag_value}
          onChange={(v) => onChange({ ...query, tag_value: String(v.value) })}
          placeholder={"Selecione o Valor"}
          defaultValue={"All"}
          allowCustomValue
        />
      </InlineField>
    </Stack>
  );
}