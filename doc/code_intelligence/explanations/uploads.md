# Precise code intelligence uploads

<style>
img.screenshot {
  display: block;
  margin: 1em auto;
  max-width: 600px;
  margin-bottom: 0.5em;
  border: 1px solid lightgrey;
  border-radius: 10px;
}

img.terminal-screenshot {
  max-width: 800px;
}
</style>

[Code intelligence indexers](../references/indexers.md) analyze source code and produce an index file. This file is [uploaded to a Sourcegraph instance](../how-to/index_other_languages.md#4-upload-lsif-data) via [Sourcegraph CLI](../../cli/index.md) for processing. Once processed, its data is available to [precise code intelligence queries](precise_code_intelligence.md).

## Lifecycle of an upload

Uploaded index files are processed asynchronously from a queue. Each upload has an attached _state_ that can change over time as work associated with that data is performed. The following diagram shows transition paths from one possible state of an upload to another.

![Upload state diagram](./diagrams/upload-states.svg)

The general happy-path for an upload is: _UPLOADING_, _QUEUED_, _PROCESSING_, then _COMPLETED_. 

Processing of a precise code intelligence index file may occur due to malformed input or due to transient errors related to the network (for example). An upload will enter the `FAILED` state on the former type of error and the `ERRORED` state. Errored uploads may be retried a number of times before moving into the `FAILED` state.

At any point, the upload record may be deleted. This can happen because the record is being replaced by a newer upload, due to [age of the upload record](../how-to/configure_data_retention.md), or due to explicit deletion by the user. Deleting a record that could be used to resolve to code intelligence queries will first move into the `DELETING` state. Moving temporarily into this state allows Sourcegraph to smoothly transition the set of code intelligence uploads that are visible for query resolution.

The change of an upload into or away from the `COMPLETED` state can be an expensive one, and when is the [repository commit graph](#repository-commit-graph) needs to be refreshed.
## Lifecycle of an upload (via UI)

After successful upload of an index file, the Sourcegraph CLI will display a URL on the target instance that shows the progress of that upload.

<img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/code-intelligence/sg-3.34/uploads/src-lsif-upload.gif" class="screenshot terminal-screenshot" alt="Uploading a precise code intelligence index via the Sourcegraph CLI">

Alternatively, users can see precise code intelligence uploads for a particular repository by navigating to the code intelligence page in the target repository's index page.

<img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/code-intelligence/sg-3.33/repository-page.png" class="screenshot" alt="Repository index page">

Administrators of a Sourcegraph instance can see a global view of precise code intelligence uploads across all repositories from the _Site Admin_ page.

<img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/code-intelligence/sg-3.34/uploads/site-admin-list.png" class="screenshot" alt="Global list of precise code intelligence uploads across all repositories">

## Repository commit graph

Sourcegraph keeps a mapping from a commit of a repository to the set of upload records that can resolve a query for that commit. When an upload record moves into or away from the `COMPLETED` state, the set of eligible uploads change and this mapping must be recalculated.

When an upload changes state, we flag the repository as needing to be updated. Then the [`worker` service](http://localhost:5080/admin/workers#codeintel-commitgraph)
will update the commit graph and unset the flag for that repository asynchronously.

While this flag is set, the repository's commit graph is considered _stale_. This simply means that there may be some upload records in a `COMPLETED` state that aren't yet being used to resolve code intelligence queries (as might be expected).

The state of a repository's commit graph can be seen in the code intelligence page in the target repository's index page.

<img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/code-intelligence/sg-3.34/uploads/list-stale-commit-graph.png" class="screenshot" alt="Stale repository commit graph notice">

Once the commit graph has updated (and no subsequent changes to that repository's uploads have occurred), the repository commit graph is no longer considered stale.

<img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/code-intelligence/sg-3.34/uploads/list-states.png" class="screenshot" alt="Up-to-date repository commit graph notice">
