import { render } from '@testing-library/react'
import React from 'react'

import { Timestamp } from './Timestamp'

describe('Timestamp', () => {
    test('mocked current time', () => expect(render(<Timestamp date="2006-01-02" />).asFragment()).toMatchSnapshot())

    test('noAbout', () => expect(render(<Timestamp date="2006-01-02" noAbout={true} />).asFragment()).toMatchSnapshot())
})
