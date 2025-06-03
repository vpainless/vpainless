/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { UUID } from './UUID';
export type Group = {
    id?: UUID;
    name?: string;
    vps?: {
        provider?: Group.provider;
        apikey?: string;
    };
};
export namespace Group {
    export enum provider {
        VULTR = 'vultr',
    }
}

