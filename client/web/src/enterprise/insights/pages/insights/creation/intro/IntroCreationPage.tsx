import React, { useEffect } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { PageHeader } from '@sourcegraph/wildcard'

import { Page } from '../../../../../../components/Page'
import { CodeInsightsIcon } from '../../../../../../insights/Icons'
import { BetaFeedbackPanel } from '../../../../components/beta-feedback-panel/BetaFeedbackPanel'

import {
    CaptureGroupInsightCard,
    ExtensionInsightsCard,
    LangStatsInsightCard,
    SearchInsightCard,
} from './cards/InsightCards'
import styles from './IntroCreationPage.module.scss'

interface IntroCreationPageProps extends TelemetryProps {}

/** Displays intro page for insights creation UI. */
export const IntroCreationPage: React.FunctionComponent<IntroCreationPageProps> = props => {
    const { telemetryService } = props

    const logCreateSearchBasedInsightClick = (): void => {
        telemetryService.log('CodeInsightsCreateSearchBasedInsightClick')
    }

    const logCreateCodeStatsInsightClick = (): void => {
        telemetryService.log('CodeInsightsCreateCodeStatsInsightClick')
    }

    const logExploreExtensionsClick = (): void => {
        telemetryService.log('CodeInsightsExploreInsightExtensionsClick')
    }

    useEffect(() => {
        telemetryService.logViewEvent('CodeInsightsCreationPage')
    }, [telemetryService])

    return (
        <Page className="container">
            <PageHeader
                annotation={<BetaFeedbackPanel />}
                path={[{ icon: CodeInsightsIcon }, { text: 'Create new code insight' }]}
                description={
                    <>
                        Insights analyze your code based on any search query.{' '}
                        <a href="https://docs.sourcegraph.com/code_insights" target="_blank" rel="noopener">
                            Learn more.
                        </a>
                    </>
                }
                className={styles.header}
            />

            <section className="pb-5">
                <h2 className="mb-3">Basic</h2>

                <div className={styles.sectionContent}>
                    <SearchInsightCard to="/insights/create/lang-stats" onClick={logCreateSearchBasedInsightClick} />

                    <CaptureGroupInsightCard
                        to="/insights/create/lang-stats"
                        onClick={logCreateSearchBasedInsightClick}
                    />

                    <LangStatsInsightCard to="/insights/create/lang-stats" onClick={logCreateCodeStatsInsightClick} />

                    <ExtensionInsightsCard
                        to="/extensions?query=category:Insights&experimental=true"
                        onClick={logExploreExtensionsClick}
                    />
                </div>

                <footer className="mt-3">
                    Not sure which insight type to choose? Learn more about the{' '}
                    <a href="https://docs.sourcegraph.com/code_insights">use cases</a>.
                </footer>
            </section>
        </Page>
    )
}
