/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { User } from '../models/User';
import type { Users } from '../models/Users';
import type { UUID } from '../models/UUID';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class UsersService {
    /**
     * Returns the logged in user info
     * Returns the information for the registered user in the system given it's credentials.
     *
     * This can be used by FE to check if credentials are correct.
     * @returns User get logged in user.
     * @throws ApiError
     */
    public static getMe(): CancelablePromise<User> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/me',
            errors: {
                401: `Unauthorized`,
                500: `Internal Server Error`,
            },
        });
    }
    /**
     * Gets a user given it's ID.
     * Returns the information for the registered user in the system given it's ID.
     * @param id ID of user
     * @returns User get user given it's id
     * @throws ApiError
     */
    public static getUser(
        id: UUID,
    ): CancelablePromise<User> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/users/{id}',
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
     * Updates a user in the system
     * This is to modify the users in the system.
     * @param id ID of user
     * @param requestBody
     * @returns User Update successful
     * @throws ApiError
     */
    public static putUser(
        id: UUID,
        requestBody: User,
    ): CancelablePromise<User> {
        return __request(OpenAPI, {
            method: 'PUT',
            url: '/users/{id}',
            path: {
                'id': id,
            },
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                400: `Bad request`,
                401: `Unauthorized`,
                404: `Not Found`,
                500: `Internal Server Error`,
            },
        });
    }
    /**
     * Lists users in the system
     * This api lists the users in the system. The returned result includes
     * all the users that the caller can view.
     *
     * We do not support pagination for now.
     * @returns Users List of users
     * @throws ApiError
     */
    public static listUsers(): CancelablePromise<Users> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/users',
            errors: {
                401: `Unauthorized`,
                500: `Internal Server Error`,
            },
        });
    }
    /**
     * Creates a user in the system
     * This api can be used for two purpose:
     *
     * 1. Registering clients in the system. They should set their usename and password.
     * Request of this kind should be anonymous, so no authorization header should be set.
     *
     * 1. To add uses to your group. The called of this request should be an admin of a group.
     * Naturally, admins should be logged in so for these kind of requests, authorization
     * header is mandatory.
     * @param requestBody
     * @returns User Username already exists
     * @throws ApiError
     */
    public static postUser(
        requestBody: User,
    ): CancelablePromise<User> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/users',
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                400: `Bad request`,
                401: `Unauthorized`,
                404: `Not Found`,
                500: `Internal Server Error`,
            },
        });
    }
}
