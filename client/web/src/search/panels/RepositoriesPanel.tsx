import classNames from 'classnames'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { Observable } from 'rxjs'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import { isRepoFilter } from '@sourcegraph/shared/src/search/query/validate'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { parseSearchURLQuery } from '..'
import { AuthenticatedUser } from '../../auth'
import { SyntaxHighlightedSearchQuery } from '../../components/SyntaxHighlightedSearchQuery'
import { EventLogResult } from '../backend'

import { EmptyPanelContainer } from './EmptyPanelContainer'
import { LoadingPanelView } from './LoadingPanelView'
import { PanelContainer } from './PanelContainer'
import { ShowMoreButton } from './ShowMoreButton'

interface Props extends TelemetryProps {
    className?: string
    authenticatedUser: AuthenticatedUser | null
    fetchRecentSearches: (userId: string, first: number) => Observable<EventLogResult | null>
}

export const RepositoriesPanel: React.FunctionComponent<Props> = ({
    className,
    authenticatedUser,
    fetchRecentSearches,
    telemetryService,
}) => {
    // Use a larger page size because not every search may have a `repo:` filter, and `repo:` filters could often
    // be duplicated. Therefore, we fetch more searches to populate this panel.
    const pageSize = 50
    const [itemsToLoad, setItemsToLoad] = useState(pageSize)

    const logRepoClicked = useCallback(() => telemetryService.log('RepositoriesPanelRepoFilterClicked'), [
        telemetryService,
    ])

    const loadingDisplay = <LoadingPanelView text="Loading recently searched repositories" />

    const emptyDisplay = (
        <EmptyPanelContainer className="text-muted">
            <small className="mb-2">
                <p className="mb-1">Recently searched repositories will be displayed here.</p>
                <p className="mb-1">
                    Search in repositories with the <strong>repo:</strong> filter:
                </p>
                <p className="mb-1">
                    <SyntaxHighlightedSearchQuery query="repo:sourcegraph/sourcegraph" />
                </p>
                <p className="mb-1">Add the code host to scope to a single repository:</p>
                <p className="mb-1">
                    <SyntaxHighlightedSearchQuery query="repo:^git\.local/my/repo$" />
                </p>
            </small>
        </EmptyPanelContainer>
    )

    const [repoFilterValues, setRepoFilterValues] = useState<string[] | null>(null)

    const searchEventLogs = useObservable(
        useMemo(() => fetchRecentSearches(authenticatedUser?.id || '', itemsToLoad), [
            fetchRecentSearches,
            authenticatedUser?.id,
            itemsToLoad,
        ])
    )

    useEffect(() => {
        if (searchEventLogs) {
            const recentlySearchedRepos = processRepositories(searchEventLogs)
            setRepoFilterValues(recentlySearchedRepos)
        }
    }, [searchEventLogs])

    useEffect(() => {
        // Only log the first load (when items to load is equal to the page size)
        if (repoFilterValues && itemsToLoad === pageSize) {
            telemetryService.log(
                'RepositoriesPanelLoaded',
                { empty: repoFilterValues.length === 0 },
                { empty: repoFilterValues.length === 0 }
            )
        }
    }, [repoFilterValues, telemetryService, itemsToLoad])

    function loadMoreItems(): void {
        setItemsToLoad(current => current + pageSize)
        telemetryService.log('RepositoriesPanelShowMoreClicked')
    }

    const contentDisplay = (
        <div className="mt-2">
            <div className="d-flex mb-1">
                <small>Search</small>
            </div>
            {repoFilterValues?.map((repoFilterValue, index) => (
                <dd key={`${repoFilterValue}-${index}`} className="text-monospace text-break">
                    <small>
                        <Link to={`/search?q=repo:${repoFilterValue}`} onClick={logRepoClicked}>
                            <SyntaxHighlightedSearchQuery query={`repo:${repoFilterValue}`} />
                        </Link>
                    </small>
                </dd>
            ))}
            {searchEventLogs?.pageInfo.hasNextPage && (
                <ShowMoreButton className="test-repositories-panel-show-more" onClick={loadMoreItems} />
            )}
        </div>
    )

    return (
        <PanelContainer
            className={classNames(className, 'repositories-panel')}
            title="Repositories"
            state={repoFilterValues ? (repoFilterValues.length > 0 ? 'populated' : 'empty') : 'loading'}
            loadingContent={loadingDisplay}
            populatedContent={contentDisplay}
            emptyContent={emptyDisplay}
        />
    )
}

function processRepositories(eventLogResult: EventLogResult): string[] | null {
    if (!eventLogResult) {
        return null
    }

    const recentlySearchedRepos: string[] = []

    for (const node of eventLogResult.nodes) {
        if (node.url) {
            const url = new URL(node.url)
            const queryFromURL = parseSearchURLQuery(url.search)
            const scannedQuery = scanSearchQuery(queryFromURL || '')
            if (scannedQuery.type === 'success') {
                for (const token of scannedQuery.term) {
                    if (isRepoFilter(token)) {
                        if (token.value && !recentlySearchedRepos.includes(token.value.value)) {
                            recentlySearchedRepos.push(token.value.value)
                        }
                    }
                }
            }
        }
    }
    return recentlySearchedRepos
}
