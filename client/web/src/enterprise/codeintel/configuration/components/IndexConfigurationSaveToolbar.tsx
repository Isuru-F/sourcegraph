import React from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { SaveToolbar, SaveToolbarProps } from '@sourcegraph/web/src/components/SaveToolbar'
import { Button } from '@sourcegraph/wildcard'

export interface IndexConfigurationSaveToolbarProps {
    loading: boolean
    inferEnabled: boolean
    onInfer?: () => void
}

export const IndexConfigurationSaveToolbar: React.FunctionComponent<
    SaveToolbarProps & IndexConfigurationSaveToolbarProps
> = ({ dirty, loading, saving, error, onSave, onDiscard, inferEnabled, onInfer, saveDiscardDisabled }) => (
    <SaveToolbar
        dirty={dirty}
        saving={saving}
        onSave={onSave}
        error={error}
        saveDiscardDisabled={saveDiscardDisabled}
        onDiscard={onDiscard}
    >
        {loading ? (
            <LoadingSpinner className="icon-inline mt-2 ml-2" />
        ) : (
            inferEnabled && (
                <Button type="button" title="Infer index configuration from HEAD" variant="link" onClick={onInfer}>
                    Infer index configuration from HEAD
                </Button>
            )
        )}
    </SaveToolbar>
)
