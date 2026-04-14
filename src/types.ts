import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';

export interface MyQuery extends DataQuery {
  det: string;
  type: string;
  namespace: string;
  tag_key: string;
  tag_value: string;
  service: string;
}

export const DEFAULT_QUERY: Partial<MyQuery> = {
  det: "Dias",
  type: "all",
  namespace: "All",
  tag_key: "All",
  tag_value: "All",
  service: "All"
};

export interface DataPoint {
  Time: number;
  Value: number;
}

export interface DataSourceResponse {
  datapoints: DataPoint[];
}

/**
 * These are options configured for each DataSource instance
 */
export interface MyDataSourceOptions extends DataSourceJsonData {
  path?: string;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface MySecureJsonData {
  UserOCID?: string;
  TenancyOCID?: string;
  Fingerprint?: string;
  Region?: string;
  PrivateKey?: string;
}
