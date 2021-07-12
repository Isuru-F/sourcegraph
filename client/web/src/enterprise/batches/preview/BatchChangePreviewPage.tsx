import { isEqual } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React, { useEffect, useMemo } from 'react'
import { delay, distinctUntilChanged, repeatWhen } from 'rxjs/operators'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { PageHeader } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../auth'
import { BatchChangesIcon } from '../../../batches/icons'
import { HeroPage } from '../../../components/HeroPage'
import { PageTitle } from '../../../components/PageTitle'
import { Description } from '../Description'
import { SupersedingBatchSpecAlert } from '../detail/SupersedingBatchSpecAlert'
import { MultiSelectContextProvider } from '../MultiSelectContext'

import { fetchBatchSpecById as _fetchBatchSpecById } from './backend'
import { BatchChangePreviewContextProvider } from './BatchChangePreviewContext'
import { BatchChangePreviewStatsBar } from './BatchChangePreviewStatsBar'
import { BatchChangePreviewProps, BatchChangePreviewTabs } from './BatchChangePreviewTabs'
import { BatchSpecInfoByline } from './BatchSpecInfoByline'
import { CreateUpdateBatchChangeAlert } from './CreateUpdateBatchChangeAlert'
import { MissingCredentialsAlert } from './MissingCredentialsAlert'

export type PreviewPageAuthenticatedUser = Pick<AuthenticatedUser, 'url' | 'displayName' | 'username' | 'email'>

export interface BatchChangePreviewPageProps extends BatchChangePreviewProps {
    /** Used for testing. */
    fetchBatchSpecById?: typeof _fetchBatchSpecById
}

export const BatchChangePreviewPage: React.FunctionComponent<BatchChangePreviewPageProps> = props => {
    const {
        batchSpecID: specID,
        history,
        authenticatedUser,
        telemetryService,
        fetchBatchSpecById = _fetchBatchSpecById,
    } = props

    const spec = useObservable(
        useMemo(
            () =>
                fetchBatchSpecById(specID).pipe(
                    repeatWhen(notifier => notifier.pipe(delay(5000))),
                    distinctUntilChanged((a, b) => isEqual(a, b))
                ),
            [specID, fetchBatchSpecById]
        )
    )

    useEffect(() => {
        telemetryService.logViewEvent('BatchChangeApplyPage')
    }, [telemetryService])

    if (spec === undefined) {
        return (
            <div className="text-center">
                <LoadingSpinner className="icon-inline mx-auto my-4" />
            </div>
        )
    }
    if (spec === null) {
        return <HeroPage icon={AlertCircleIcon} title="Batch spec not found" />
    }

    return (
        <BatchChangePreviewContextProvider>
            <MultiSelectContextProvider>
                <div className="pb-5">
                    <PageTitle title="Apply batch spec" />
                    <PageHeader
                        path={[
                            {
                                icon: BatchChangesIcon,
                                to: '/batch-changes',
                            },
                            { to: `${spec.namespace.url}/batch-changes`, text: spec.namespace.namespaceName },
                            { text: spec.description.name },
                        ]}
                        byline={<BatchSpecInfoByline createdAt={spec.createdAt} creator={spec.creator} />}
                        headingElement="h2"
                        className="test-batch-change-apply-page mb-3"
                    />
                    <MissingCredentialsAlert
                        authenticatedUser={authenticatedUser}
                        viewerBatchChangesCodeHosts={spec.viewerBatchChangesCodeHosts}
                    />
                    <SupersedingBatchSpecAlert spec={spec.supersedingBatchSpec} />
                    <BatchChangePreviewStatsBar batchSpec={spec} />
                    <CreateUpdateBatchChangeAlert
                        specID={spec.id}
                        batchChange={spec.appliesToBatchChange}
                        toBeArchived={spec.applyPreview.stats.archive}
                        showPublishUI={spec.applyPreview.stats.uiPublished > 0}
                        viewerCanAdminister={spec.viewerCanAdminister}
                        history={history}
                        telemetryService={telemetryService}
                    />
                    <Description description={spec.description.description} />
                    <BatchChangePreviewTabs spec={spec} {...props} />
                </div>
            </MultiSelectContextProvider>
        </BatchChangePreviewContextProvider>
    )
}
