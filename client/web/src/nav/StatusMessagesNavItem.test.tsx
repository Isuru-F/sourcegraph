import { render } from '@testing-library/react'
import { createMemoryHistory } from 'history'
import React from 'react'
import { of, Observable } from 'rxjs'

import { StatusMessagesResult, StatusMessageFields } from '../graphql-operations'
import { SourcegraphContext } from '../jscontext'

import { StatusMessagesNavItem } from './StatusMessagesNavItem'

describe('StatusMessagesNavItem', () => {
    const user = {
        id: 'VXNlcjox',
        username: 'user',
        isSiteAdmin: false,
    }

    let oldContext: SourcegraphContext & Mocha.SuiteFunction

    beforeEach(() => {
        oldContext = window.context
        window.context = {} as SourcegraphContext & Mocha.SuiteFunction
    })
    afterEach(() => {
        window.context = oldContext
    })

    test('no messages', () => {
        const fetchMessages = (): Observable<StatusMessagesResult['statusMessages']> => of([])
        expect(
            render(
                <StatusMessagesNavItem fetchMessages={fetchMessages} user={user} history={createMemoryHistory()} />
            ).asFragment()
        ).toMatchSnapshot()
    })

    describe('one CloningProgress message', () => {
        const message: StatusMessageFields = {
            type: 'CloningProgress',
            message: 'Currently cloning repositories...',
        }

        const fetchMessages = () => of([message])
        test('as non-site admin', () => {
            expect(
                render(
                    <StatusMessagesNavItem fetchMessages={fetchMessages} user={user} history={createMemoryHistory()} />
                ).asFragment()
            ).toMatchSnapshot()
        })

        test('as site admin', () => {
            expect(
                render(
                    <StatusMessagesNavItem
                        fetchMessages={fetchMessages}
                        user={{ ...user, isSiteAdmin: true }}
                        history={createMemoryHistory()}
                    />
                ).asFragment()
            ).toMatchSnapshot()
        })
    })

    describe('one ExternalServiceSyncError message', () => {
        const message: StatusMessageFields = {
            type: 'ExternalServiceSyncError',
            message: 'failed to list organization kubernetes repos: request returned status 404: Not Found',
            externalService: {
                id: 'abcd',
                displayName: 'GitHub.com',
            },
        }

        const fetchMessages = () => of([message])
        test('as non-site admin', () => {
            expect(
                render(
                    <StatusMessagesNavItem fetchMessages={fetchMessages} user={user} history={createMemoryHistory()} />
                ).asFragment()
            ).toMatchSnapshot()
        })

        test('as site admin', () => {
            expect(
                render(
                    <StatusMessagesNavItem
                        fetchMessages={fetchMessages}
                        user={{ ...user, isSiteAdmin: true }}
                        history={createMemoryHistory()}
                    />
                ).asFragment()
            ).toMatchSnapshot()
        })
    })

    describe('one SyncError message', () => {
        const message: StatusMessageFields = {
            type: 'SyncError',
            message: 'syncer.sync.store.upsert-repos: pg: unique constraint foobar',
        }

        const fetchMessages = () => of([message])
        test('as non-site admin', () => {
            expect(
                render(
                    <StatusMessagesNavItem fetchMessages={fetchMessages} user={user} history={createMemoryHistory()} />
                ).asFragment()
            ).toMatchSnapshot()
        })

        test('as site admin', () => {
            expect(
                render(
                    <StatusMessagesNavItem
                        fetchMessages={fetchMessages}
                        user={{ ...user, isSiteAdmin: true }}
                        history={createMemoryHistory()}
                    />
                ).asFragment()
            ).toMatchSnapshot()
        })
    })
})
