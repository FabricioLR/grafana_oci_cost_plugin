import React, { ChangeEvent } from 'react';
import { InlineField, Input, SecretInput } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { MyDataSourceOptions, MySecureJsonData } from '../types';

interface Props extends DataSourcePluginOptionsEditorProps<MyDataSourceOptions, MySecureJsonData> {}

export function ConfigEditor(props: Props) {
  const { onOptionsChange, options } = props;
  const { secureJsonFields, secureJsonData } = options;

  const onUserOcidChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      secureJsonData: {
        ...options.secureJsonData,
        UserOCID: event.target.value,
      },
    });
  };

  const onResetUserOcid = () => {
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        UserOCID: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        UserOCID: '',
      },
    });
  };

  const onTenancyOcidChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      secureJsonData: {
        ...options.secureJsonData,
        TenancyOCID: event.target.value,
      },
    });
  };

  const onResetTenancyOcid = () => {
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        TenancyOCID: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        TenancyOCID: '',
      },
    });
  };

  const onFingerprintChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      secureJsonData: {
        ...options.secureJsonData,
        Fingerprint: event.target.value,
      },
    });
  };

  const onResetFingerprint = () => {
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        Fingerprint: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        Fingerprint: '',
      },
    });
  };

  const onRegionChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      secureJsonData: {
        ...options.secureJsonData,
        Region: event.target.value,
      },
    });
  };

  const onResetRegion = () => {
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        Region: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        Region: '',
      },
    });
  };

  const onPrivateKeyChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      secureJsonData: {
        ...options.secureJsonData,
        PrivateKey: event.target.value,
      },
    });
  };

  const onResetPrivateKey = () => {
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        PrivateKey: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        PrivateKey: '',
      },
    });
  };

  return (
    <>
      <InlineField label="User OCID" labelWidth={25} interactive tooltip={'Secure json field (backend only)'}>
        <SecretInput
          required
          id="config-editor-user-ocid"
          isConfigured={secureJsonFields.UserOCID}
          value={secureJsonData?.UserOCID}
          placeholder="Enter your OCI User Ocid"
          width={150}
          onReset={onResetUserOcid}
          onChange={onUserOcidChange}
        />
      </InlineField>
      <InlineField label="Tenancy OCID" labelWidth={25} interactive tooltip={'Secure json field (backend only)'}>
        <SecretInput
          required
          id="config-editor-tenancy-ocid"
          isConfigured={secureJsonFields.TenancyOCID}
          value={secureJsonData?.TenancyOCID}
          placeholder="Enter your OCI Tenancy Ocid"
          width={150}
          onReset={onResetTenancyOcid}
          onChange={onTenancyOcidChange}
        />
      </InlineField>
      <InlineField label="Fingerprint" labelWidth={25} interactive tooltip={'Secure json field (backend only)'}>
        <SecretInput
          required
          id="config-editor-Fingerprint"
          isConfigured={secureJsonFields.Fingerprint}
          value={secureJsonData?.Fingerprint}
          placeholder="Enter your Key Fingerprint"
          width={150}
          onReset={onResetFingerprint}
          onChange={onFingerprintChange}
        />
      </InlineField>
      <InlineField label="Region" labelWidth={25} interactive tooltip={'Secure json field (backend only)'}>
        <SecretInput
          required
          id="config-editor-Region"
          isConfigured={secureJsonFields.Region}
          value={secureJsonData?.Region}
          placeholder="Enter your OCI Region"
          width={150}
          onReset={onResetRegion}
          onChange={onRegionChange}
        />
      </InlineField>
      <InlineField label="Private Key" labelWidth={25} interactive tooltip={'Secure json field (backend only)'}>
        <SecretInput
          required
          id="config-editor-private-key"
          isConfigured={secureJsonFields.PrivateKey}
          value={secureJsonData?.PrivateKey}
          placeholder="Enter your OCI Private Key"
          width={150}
          onReset={onResetPrivateKey}
          onChange={onPrivateKeyChange}
        />
      </InlineField>
    </>
  );
}
