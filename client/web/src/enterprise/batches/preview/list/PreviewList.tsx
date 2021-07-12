import * as H from 'history'
import MagnifyIcon from 'mdi-react/MagnifyIcon'
import React, { useCallback, useContext } from 'react'
import { tap } from 'rxjs/operators'

import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Container } from '@sourcegraph/wildcard'

import { FilteredConnection, FilteredConnectionQueryArguments } from '../../../../components/FilteredConnection'
import { ChangesetApplyPreviewFields, Scalars } from '../../../../graphql-operations'
import { MultiSelectContext } from '../../MultiSelectContext'
import { BatchChangePreviewContext } from '../BatchChangePreviewContext'
import { PreviewPageAuthenticatedUser } from '../BatchChangePreviewPage'

import { queryChangesetApplyPreview as _queryChangesetApplyPreview, queryChangesetSpecFileDiffs } from './backend'
import { ChangesetApplyPreviewNode, ChangesetApplyPreviewNodeProps } from './ChangesetApplyPreviewNode'
import { EmptyPreviewListElement } from './EmptyPreviewListElement'
import { PreviewFilterRow } from './PreviewFilterRow'
import styles from './PreviewList.module.scss'
import { PreviewListHeader, PreviewListHeaderProps } from './PreviewListHeader'

interface Props extends ThemeProps {
    batchSpecID: Scalars['ID']
    history: H.History
    location: H.Location
    authenticatedUser: PreviewPageAuthenticatedUser

    selectionEnabled: boolean

    /** For testing only. */
    queryChangesetApplyPreview?: typeof _queryChangesetApplyPreview
    /** For testing only. */
    queryChangesetSpecFileDiffs?: typeof queryChangesetSpecFileDiffs
    /** Expand changeset descriptions, for testing only. */
    expandChangesetDescriptions?: boolean
}

/**
 * A list of a batch spec's preview nodes.
 */
export const PreviewList: React.FunctionComponent<Props> = ({
    batchSpecID,
    history,
    location,
    authenticatedUser,
    isLightTheme,

    selectionEnabled,

    queryChangesetApplyPreview = _queryChangesetApplyPreview,
    queryChangesetSpecFileDiffs,
    expandChangesetDescriptions,
}) => {
    const { filters, setPagination } = useContext(BatchChangePreviewContext)
    const { setVisible: onLoad } = useContext(MultiSelectContext)

    const queryChangesetApplyPreviewConnection = useCallback(
        (args: FilteredConnectionQueryArguments) => {
            const pagination = { after: args.after ?? null, first: args.first ?? null }
            setPagination(pagination)

            return queryChangesetApplyPreview({ batchSpec: batchSpecID, ...filters, ...pagination }).pipe(
                tap(connection => {
                    onLoad(
                        connection.nodes
                            .map(node => {
                                if (node.__typename === 'HiddenChangesetApplyPreview') {
                                    return undefined
                                }
                                if (node.targets.__typename === 'VisibleApplyPreviewTargetsDetach') {
                                    return undefined
                                }
                                return node.targets.changesetSpec.id
                            })
                            .filter((id): id is string => id !== undefined)
                    )
                })
            )
        },
        [setPagination, queryChangesetApplyPreview, batchSpecID, filters, onLoad]
    )

    return (
        <Container>
            <PreviewFilterRow history={history} location={location} />
            <FilteredConnection<
                ChangesetApplyPreviewFields,
                Omit<ChangesetApplyPreviewNodeProps, 'node'>,
                PreviewListHeaderProps
            >
                className="mt-2"
                nodeComponent={ChangesetApplyPreviewNode}
                nodeComponentProps={{
                    isLightTheme,
                    history,
                    location,
                    authenticatedUser,
                    queryChangesetSpecFileDiffs,
                    expandChangesetDescriptions,
                    selectionEnabled,
                }}
                queryConnection={queryChangesetApplyPreviewConnection}
                hideSearch={true}
                defaultFirst={15}
                noun="changeset"
                pluralNoun="changesets"
                history={history}
                location={location}
                useURLQuery={true}
                listComponent="div"
                listClassName={styles.previewListGrid}
                headComponent={PreviewListHeader}
                headComponentProps={{
                    selectionEnabled,
                }}
                cursorPaging={true}
                noSummaryIfAllNodesVisible={true}
                emptyElement={
                    filters.search || filters.currentState || filters.action ? (
                        <EmptyPreviewSearchElement />
                    ) : (
                        <EmptyPreviewListElement />
                    )
                }
            />
        </Container>
    )
}

const EmptyPreviewSearchElement: React.FunctionComponent<{}> = () => (
    <div className="text-muted row w-100">
        <div className="col-12 text-center">
            <MagnifyIcon className="icon" />
            <div className="pt-2">No changesets matched the search.</div>
        </div>
    </div>
)
