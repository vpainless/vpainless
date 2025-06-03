/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { Instance } from '../models/Instance';
import type { UUID } from '../models/UUID';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class InstancesService {
    /**
     * Gets an instance given it's ID.
     * Returns the information for an instance in the system given it's ID.
     * @param id ID of pet to use
     * @returns Instance get instance given it's id
     * @throws ApiError
     */
    public static getInstance(
        id: UUID,
    ): CancelablePromise<Instance> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/instances/{id}',
            path: {
                'id': id,
            },
            errors: {
                401: `Unauthorized`,
                404: `Not Found`,
                500: `Internal Server Error`,
            },
        });
    }
    /**
     * Deletes an instance given it's ID.
     * Deletes the instance in the system. It also deletes the instance created by the provided.
     *
     * This should be used by clients before renewing instances, to save costs.
     * @param id ID of pet to use
     * @returns void
     * @throws ApiError
     */
    public static deleteInstance(
        id: UUID,
    ): CancelablePromise<void> {
        return __request(OpenAPI, {
            method: 'DELETE',
            url: '/instances/{id}',
            path: {
                'id': id,
            },
            errors: {
                401: `Unauthorized`,
                404: `Not Found`,
                500: `Internal Server Error`,
            },
        });
    }
    /**
     * List the instances
     * Using this, users can list the instances they can view. clients will see
     * the instances associated to them. Group admins can list all the instances
     * associated to their clients.
     * @returns Instance Listed instances
     * @throws ApiError
     */
    public static listInstances(): CancelablePromise<Array<Instance>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/instances',
            errors: {
                401: `Unauthorized`,
                500: `Internal Server Error`,
            },
        });
    }
    /**
     * Creates an instance in the system
     * Using this, users can create an instance in the system. Instance will be created
     * using the default values of the group they are part of.
     * @returns Instance Instance existed already.
     * @throws ApiError
     */
    public static postInstance(): CancelablePromise<Instance> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/instances',
            errors: {
                400: `Bad request`,
                401: `Unauthorized`,
                500: `Internal Server Error`,
            },
        });
    }
}
