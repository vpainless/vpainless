/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { UUID } from './UUID';
export type User = {
    id?: UUID;
    username?: string;
    password?: string;
    group_id?: UUID;
    role?: User.role;
};
export namespace User {
    export enum role {
        CLIENT = 'client',
        ADMIN = 'admin',
    }
}

