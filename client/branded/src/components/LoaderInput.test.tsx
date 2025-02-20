import { render } from '@testing-library/react'
import React from 'react'

import { LoaderInput } from './LoaderInput'

jest.mock('@sourcegraph/react-loading-spinner', () => ({ LoadingSpinner: 'LoadingSpinner' }))

describe('LoaderInput', () => {
    it('should render a loading spinner when loading prop is true', () => {
        expect(
            render(
                <LoaderInput loading={true}>
                    <input type="text" />
                </LoaderInput>
            ).asFragment()
        ).toMatchSnapshot()
    })

    it('should not render a loading spinner when loading prop is false', () => {
        expect(
            render(
                <LoaderInput loading={false}>
                    <input type="text" />
                </LoaderInput>
            ).asFragment()
        ).toMatchSnapshot()
    })
})
