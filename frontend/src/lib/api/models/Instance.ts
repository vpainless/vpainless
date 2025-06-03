/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { UUID } from './UUID';
export type Instance = {
    id?: UUID;
    owner?: UUID;
    ip?: string;
    connection_string?: string;
    status?: Instance.status;
};
export namespace Instance {
    export enum status {
        UNKNOWN = 'unknown',
        OFF = 'off',
        INITIALIZING = 'initializing',
        OK = 'ok',
    }
}

