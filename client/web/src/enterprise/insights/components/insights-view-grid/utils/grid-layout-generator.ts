import { Layout, Layouts as ReactGridLayouts } from 'react-grid-layout'

import {
    BreakpointName,
    BREAKPOINTS_NAMES,
    COLUMNS,
    DEFAULT_HEIGHT,
    DEFAULT_ITEMS_PER_ROW,
    MIN_WIDTHS,
} from '../../../../../views'
import { MINIMAL_SERIES_FOR_ASIDE_LEGEND } from '../../../../../views/components/view/content/chart-view-content/charts/line/constants'
import { Insight, isSearchBasedInsight } from '../../../core/types'

/**
 * Custom Code Insight Grid layout generator. For different screens (xs, sm, md, lg) it
 * generates different initial layouts. See examples below
 *
 * <pre>
 * Large break points (lg)                         Mid size and small breakpoints (xs, sm, md)
 * ┌────────────┐ ┌────────────┐ ┌────────────┐    ┌────────────┐ ┌────────────┐ ┌────────────┐
 * │▪▪▪▪▪▪▪▪▪   │ │▪▪▪▪▪       │ │▪▪▪▪▪▪▪▪▪   │    │▪▪▪▪▪▪▪▪▪   │ │▪▪▪▪▪       │ │▪▪▪▪▪▪▪▪▪   │
 * │            │ │            │ │            │    │            │ │            │ │            │
 * │            │ │            │ │            │    │            │ │            │ │            │
 * │            │ │            │ │            │    │            │ │            │ │            │
 * │           ◿│ │           ◿│ │           ◿│    │           ◿│ │           ◿│ │           ◿│
 * └────────────┘ └────────────┘ └────────────┘    └────────────┘ └────────────┘ └────────────┘
 * ┌────────────────────┐┌────────────────────┐    ┌────────────┐ ┌────────────┐ ┌────────────┐
 * │■■■■■■■■■■■■■       ││■■■■■■■■■           │    │■■■■■■■■■■■■│ │▪▪▪▪▪▪▪▪▪   │ │▪▪▪▪▪       │
 * │                    ││                    │    │            │ │            │ │            │
 * │ Insight with 3 and ││Insight with 3 and  │    │ Insight    │ │            │ │            │
 * │ more series        ││more series         │    │ with 3 and │ │            │ │            │
 * │                   ◿││                   ◿│    │ more       │ │           ◿│ │           ◿│
 * └────────────────────┘└────────────────────┘    │            │ └────────────┘ └────────────┘
 * ┌────────────┐ ┌────────────┐                   │            │ ┌────────────┐ ┌────────────┐
 * │▪▪▪▪▪▪▪▪▪   │ │▪▪▪▪▪       │                   │            │ │▪▪▪▪▪▪▪▪▪   │ │▪▪▪▪▪▪▪▪▪   │
 * │            │ │            │                   └────────────┘ │            │ │            │
 * │            │ │            │                                  │            │ │            │
 * │            │ │            │                                  │            │ │            │
 * │           ◿│ │           ◿│                                  │           ◿│ │           ◿│
 * └────────────┘ └────────────┘                                  └────────────┘ └────────────┘
 * </pre>
 */
export const insightLayoutGenerator = (insights: Insight[]): ReactGridLayouts => {
    return Object.fromEntries(
        BREAKPOINTS_NAMES.map(breakpointName => [breakpointName, generateLayout(breakpointName)] as const)
    )

    function generateLayout(breakpointName: BreakpointName): Layout[] {
        switch (breakpointName) {
            case 'xs':
            case 'sm':
            case 'md': {
                return insights.map((insight, index) => {
                    const isManySeriesChart =
                        isSearchBasedInsight(insight) && insight.series.length > MINIMAL_SERIES_FOR_ASIDE_LEGEND
                    const width = COLUMNS[breakpointName] / DEFAULT_ITEMS_PER_ROW[breakpointName]
                    return {
                        i: insight.id,
                        // Increase height of chart block if view has many data series
                        h: isManySeriesChart ? DEFAULT_HEIGHT * insight.series.length * 0.3 : DEFAULT_HEIGHT,
                        w: width,
                        x: (index * width) % COLUMNS[breakpointName],
                        y: Math.floor((index * width) / COLUMNS[breakpointName]),
                        minW: MIN_WIDTHS[breakpointName],
                        minH: isManySeriesChart ? DEFAULT_HEIGHT * insight.series.length * 0.15 : DEFAULT_HEIGHT,
                    }
                })
            }

            case 'lg': {
                return insights
                    .reduce<Layout[][]>(
                        (grid, insight) => {
                            const isManySeriesChart =
                                isSearchBasedInsight(insight) && insight.series.length > MINIMAL_SERIES_FOR_ASIDE_LEGEND
                            const itemsPerRow = isManySeriesChart ? 2 : DEFAULT_ITEMS_PER_ROW[breakpointName]
                            const columnsPerRow = COLUMNS[breakpointName]
                            const width = columnsPerRow / itemsPerRow
                            const lastRow = grid[grid.length - 1]
                            const lastRowCurrentWidth = lastRow.reduce((sumWidth, element) => sumWidth + element.w, 0)

                            // Move element on new line (row)
                            if (lastRowCurrentWidth + width > columnsPerRow) {
                                // Adjust elements width on the same row if no more elements don't
                                // fit in this row
                                for (const [index, element] of lastRow.entries()) {
                                    element.w = columnsPerRow / lastRow.length
                                    element.x = (index * columnsPerRow) / lastRow.length
                                }

                                // Create new row
                                grid.push([
                                    {
                                        i: insight.id,
                                        h: DEFAULT_HEIGHT,
                                        w: width,
                                        x: 0,
                                        y: grid.length,
                                        minW: MIN_WIDTHS[breakpointName],
                                        minH: DEFAULT_HEIGHT,
                                    },
                                ])
                            } else {
                                // Add another element to the last row of the grid
                                lastRow.push({
                                    i: insight.id,
                                    h: DEFAULT_HEIGHT,
                                    w: width,
                                    x: lastRowCurrentWidth,
                                    y: grid.length - 1,
                                    minW: MIN_WIDTHS[breakpointName],
                                    minH: DEFAULT_HEIGHT,
                                })
                            }

                            return grid
                        },
                        [[]]
                    )
                    .flat()
            }
        }
    }
}
