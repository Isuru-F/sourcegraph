import classNames from 'classnames'
import Check from 'mdi-react/CheckIcon'
import Info from 'mdi-react/InfoCircleOutlineIcon'
import React from 'react'
import { noop } from 'rxjs'

import { Button } from '@sourcegraph/wildcard/src'

import { FormInput } from '../../../../../../components/form/form-input/FormInput'
import { useField } from '../../../../../../components/form/hooks/useField'
import { useForm } from '../../../../../../components/form/hooks/useForm'
import { createRequiredValidator } from '../../../../../../components/form/validators'
import { SearchBasedInsightSeries } from '../../../../../../core/types/insight/search-insight'
import { DEFAULT_ACTIVE_COLOR, FormColorInput } from '../form-color-input/FormColorInput'

import { DataSeriesQueryField } from './DataSeriesQueryField'
import styles from './FormSeriesInput.module.scss'

const requiredNameField = createRequiredValidator('Name is a required field for data series.')
const validQuery = createRequiredValidator('Query is a required field for data series.')

interface FormSeriesInputProps {
    id: string | null

    /** Series index. */
    index: number

    /**
     * This prop represents the case whenever the edit insight UI page
     * deals with backend insight. We need to disable our search insight
     * query field since our backend insight can't update BE data according
     * to the latest insight configuration.
     */
    isSearchQueryDisabled: boolean

    /**
     * Show all validation error of all fields within the form.
     */
    showValidationErrorsOnMount?: boolean
    /** Name of series. */
    name?: string
    /** Query value of series. */
    query?: string
    /** Color value for line chart. (series) */
    stroke?: string
    /** Enable autofocus behavior of first input of form. */
    autofocus?: boolean
    /** Enable cancel button. */
    cancel?: boolean
    /** Custom class name for root element of form series. */
    className?: string
    /** On submit handler of series form. */
    onSubmit?: (series: SearchBasedInsightSeries) => void
    /** On cancel handler. */
    onCancel?: () => void
    /** Change handler in order to listen last values of series form. */
    onChange?: (formValues: SearchBasedInsightSeries, valid: boolean) => void
}

const SearchQueryChecks: React.FunctionComponent = () => (
    <div className={classNames(styles.formSeriesInput)}>
        <ul className={classNames(['mt-4 text-muted', styles.formSeriesInputSeriesCheck])}>
            <li>
                <Check size={16} className="text-success" /> Contains a properly formatted regular expression with at
                least one capture group
            </li>
            <li>
                <Check size={16} className="text-success" /> Does not contain boolean operator <code>AND</code> and{' '}
                <code>OR</code> (regular expression boolean operators can still be used)
            </li>
            <li>
                <Check size={16} className="text-success" /> Does not contain <code>patternType:literal</code> and{' '}
                <code>patternType:structural</code>
            </li>
            <li>
                <Check size={16} className="text-success" /> The capture group matches file contents (not{' '}
                <code>repo</code> or <code>file</code>)
            </li>
            <li>
                <Check size={16} className="text-success" /> Does not contain <code>commit</code> or <code>diff</code>{' '}
                search
            </li>
            <li>
                <Check size={16} className="text-success" /> Does not contain the <code>repo:</code> filter as it will
                be added automatically if needed
            </li>
        </ul>
        <p className="mt-4 text-muted">
            Tip: use <code>archived:no</code> or <code>fork:no</code> to exclude results from archived or forked
            repositories. Explore <a href="#">example queries</a> and learn more about{' '}
            <a href="#">automatically generated data series</a>.
        </p>
        <p className="mt-4 text-muted">
            <Info size={16} /> <b>Name</b> and <b>color</b> of each data seris will be generated automatically. Chart
            will display <b>up to 20</b> data series.
        </p>
    </div>
)

/** Displays form series input (three field - name field, query field and color picker). */
export const FormSeriesInput: React.FunctionComponent<FormSeriesInputProps> = props => {
    const {
        id,
        index,
        isSearchQueryDisabled,
        showValidationErrorsOnMount = false,
        name,
        query,
        stroke: color,
        className,
        cancel = false,
        autofocus = true,
        onCancel = noop,
        onSubmit = noop,
        onChange = noop,
    } = props

    const hasNameControlledValue = !!name
    const hasQueryControlledValue = !!query

    const { formAPI, handleSubmit, ref } = useForm({
        touched: showValidationErrorsOnMount,
        initialValues: {
            seriesName: name ?? '',
            seriesQuery: query ?? '',
            seriesColor: color ?? DEFAULT_ACTIVE_COLOR,
        },
        onSubmit: values =>
            onSubmit({
                id,
                name: values.seriesName,
                query: values.seriesQuery,
                stroke: values.seriesColor,
            }),
        onChange: event => {
            const { values } = event

            onChange(
                {
                    id,
                    name: values.seriesName,
                    query: values.seriesQuery,
                    stroke: values.seriesColor,
                },
                event.valid
            )
        },
    })

    const nameField = useField({
        name: 'seriesName',
        formApi: formAPI,
        validators: { sync: requiredNameField },
    })

    const queryField = useField({
        name: 'seriesQuery',
        formApi: formAPI,
        validators: { sync: validQuery },
        disabled: isSearchQueryDisabled,
    })

    const colorField = useField({
        name: 'seriesColor',
        formApi: formAPI,
    })

    return (
        <div data-testid="series-form" ref={ref} className={classNames('d-flex flex-column', className)}>
            <FormInput
                title="Name"
                required={true}
                autoFocus={autofocus}
                placeholder="Example: Function component"
                description="Name shown in the legend and tooltip"
                valid={(hasNameControlledValue || nameField.meta.touched) && nameField.meta.validState === 'VALID'}
                error={nameField.meta.touched && nameField.meta.error}
                {...nameField.input}
            />

            <FormInput
                title="Search query"
                required={true}
                as={DataSeriesQueryField}
                placeholder="Example: patternType:regexp const\s\w+:\s(React\.)?FunctionComponent"
                description={
                    <span>
                        {!isSearchQueryDisabled ? (
                            <SearchQueryChecks />
                        ) : (
                            <>
                                We don't yet allow editing queries for insights over all repos. To change the query,
                                make a new insight. This is a known{' '}
                                <a
                                    href="https://docs.sourcegraph.com/code_insights/explanations/current_limitations_of_code_insights"
                                    target="_blank"
                                    rel="noopener noreferrer"
                                >
                                    beta limitation
                                </a>
                            </>
                        )}
                    </span>
                }
                valid={(hasQueryControlledValue || queryField.meta.touched) && queryField.meta.validState === 'VALID'}
                error={queryField.meta.touched && queryField.meta.error}
                className="mt-4"
                {...queryField.input}
            />

            <FormColorInput
                name={`color group of ${index} series`}
                title="Color"
                className="mt-4"
                value={colorField.input.value}
                onChange={colorField.input.onChange}
            />

            <div className="mt-4">
                <Button
                    aria-label="Submit button for data series"
                    type="button"
                    variant="secondary"
                    onClick={handleSubmit}
                >
                    Done
                </Button>

                {cancel && (
                    <Button type="button" onClick={onCancel} variant="secondary" outline={true} className="ml-2">
                        Cancel
                    </Button>
                )}
            </div>
        </div>
    )
}
