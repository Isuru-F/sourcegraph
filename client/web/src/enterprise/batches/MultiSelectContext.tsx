import React, { useCallback, useState } from 'react'

/**
 * The current selection state: either a set of IDs or "all", in which case all
 * possible IDs will be considered selected.
 *
 * Note that there is no special case for "visible": when all visible items are
 * selected and a new page is loaded, the expectation is that those new items
 * will not be selected by default.
 */
export type MultiSelectContextSelected = Set<string> | 'all'

export interface MultiSelectContextState {
    // State fields. These must not be mutated other than through the mutator
    // functions below.
    selected: MultiSelectContextSelected
    visible: Set<string>

    // General state mutators to select and deselect items.
    deselectAll: () => void
    deselectVisible: () => void
    deselectSingle: (id: string) => void
    selectAll: () => void
    selectVisible: () => void
    selectSingle: (id: string) => void

    // Sets the current set of visible IDs. This needs to happen in a single
    // call to avoid unnecessary re-renders: consumers are responsible for
    // aggregating the existing state from visible if required (for example, if
    // pagination is being performed by appending to the existing list in an
    // infinite scrolling style approach).
    setVisible: (ids: string[]) => void
}

// eslint-disable @typescript-eslint/no-unused-vars
const defaultState = (): MultiSelectContextState => ({
    selected: new Set(),
    visible: new Set(),
    deselectAll: () => {},
    deselectVisible: () => {},
    deselectSingle: (id: string) => {},
    selectAll: () => {},
    selectVisible: () => {},
    selectSingle: (id: string) => {},
    setVisible: (ids: string[]) => {},
})
// eslint-enable @typescript-eslint/no-unused-vars

/**
 * MultiSelectContext is a context that tracks which checkboxes in a paginated
 * list have been selected, providing options to select visible and select all.
 * Options are tracked by opaque string IDs.
 *
 * Use MultiSelectContextProvider to instantiate a MultiSelectContext: this will
 * set up the appropriate internal state.
 */
export const MultiSelectContext = React.createContext<MultiSelectContextState>(defaultState())

export const MultiSelectContextProvider: React.FunctionComponent<{}> = ({ children }) => {
    // Set up state and callbacks for the visible items.
    const [visible, setVisibleInternal] = useState<Set<string>>(new Set())
    const setVisible = useCallback((ids: string[]) => {
        setVisibleInternal(new Set(ids))
    }, [])

    const [selected, setSelected] = useState<MultiSelectContextSelected>(new Set())
    const selectAll = useCallback(() => setSelected('all'), [setSelected])
    const deselectAll = useCallback(() => setSelected(new Set()), [setSelected])

    const selectVisible = useCallback(() => {
        // If all items are currently selected, all visible items are therefore
        // selected by definition, and we don't need to do anything.
        if (selected === 'all') {
            return
        }

        // Otherwise, we can merge the visible items with any previously
        // selected items.
        setSelected(new Set([...visible, ...selected]))
    }, [selected, visible])

    const deselectVisible = useCallback(() => {
        // If all items are currently selected, there isn't a sensible way to
        // say "except for this specific subset" within the current data model.
        // Consumers should be careful not to allow deselectVisible to occur
        // when all items are selected.
        if (selected === 'all') {
            return
        }

        // Otherwise, we remove the items and create a new set.
        setSelected(new Set([...selected].filter(id => !visible.has(id))))
    }, [selected, visible])

    const selectSingle = useCallback(
        (id: string) => {
            const updated = new Set(selected)
            updated.add(id)

            setSelected(updated)
        },
        [selected]
    )

    const deselectSingle = useCallback(
        (id: string) => {
            const updated = new Set(selected)
            updated.delete(id)

            setSelected(updated)
        },
        [selected]
    )

    return (
        <MultiSelectContext.Provider
            value={{
                selected,
                visible,
                deselectAll,
                deselectVisible,
                deselectSingle,
                selectAll,
                selectVisible,
                selectSingle,
                setVisible,
            }}
        >
            {children}
        </MultiSelectContext.Provider>
    )
}
