# Batch Changes site admin configuration reference

[Batch Changes](../../batch_changes/index.md) is generally configured through the same [site configuration](site_config.md) and [code host configuration](../external_service/index.md) as the rest of Sourcegraph. However, Batch Changes features may require specific configuration, and those are documented here.

## Rollout windows

By default, Sourcegraph attempts to reconcile (create, update, or close) changesets as quickly as the rate limits on the code host allow. This can result in CI systems being overwhelmed if hundreds or thousands of changesets are being handled as part of a single batch change.

Configuring rollout windows allows changesets to be created and updated at a slower or faster rate based on the time of day and/or the day of the week. These windows are applied to changesets across all code hosts.

Rollout windows are configured through the `batchChanges.rolloutWindows` [site configuration option](site_config.md). If specified, this option contains an array of rollout window objects that are used to schedule changesets. The format of these objects [is given below](#rollout-window-object).

### Behavior

When rollout windows are enabled, changesets will initially enter a **Scheduled** state when their batch change is applied. Hovering or tapping on the changeset's state icon will provide an estimate of when the changeset will be reconciled.

To restore the default behavior, you can either delete the `batchChanges.rolloutWindows` option, or set it to `null`.

Or, to put it another way:

| `batchChanges.rolloutWindows` configuration | Behavior |
|---------------------------------------------|-----------|
| Omitted, or set to `null`                   | Changesets will be reconciled as fast as the code host allows; essentially the same as setting a single `{"rate": "unlimited"}` window. |
| Set to an array (even if empty)             | Changesets will be reconciled using the rate limit in the current window using [the leaky bucket behavior described below](#leaky-bucket-rate-limiting). If no window covers the current period, then no changesets will be reconciled until a window with a non-zero [`rate`](#rate) opens. |
| Any other value                             | The configuration is invalid, and an error will appear. |

#### Leaky bucket rate limiting

Rate limiting uses the [leaky bucket algorithm](https://en.wikipedia.org/wiki/Leaky_bucket) to smooth bursts in reconciliations.

Practically speaking, this means that the given rate can be thought of more as an average than as a simple resource allocation. If there are always changesets in the queue, a rate of `10/hour` means that a changeset will be reconciled approximately every six minutes, rather than ten changesets being simultaneously reconciled at the start of each hour.

### Rollout window object

A rollout window is a JSON object that looks as follows:

```json
{
  "rate": "10/hour",
  "days": ["saturday", "sunday"],
  "start": "06:00",
  "end": "20:00"
}
```

All fields are optional except for `rate`, and are described below in more detail. All times and days are handled in UTC.

In the event multiple windows overlap, the last defined window will be used.

#### `rate`

`rate` describes the rate at which changesets will be reconciled. This may be expressed in one of the following ways:

* The string `unlimited`, in which case no limit will be applied for this window, or
* A string in the format `N/UNIT`, where `N` is a number and `UNIT` is one of `second`, `minute`, or `hour`; for example, `10/hour` would allow 10 changesets to be reconciled per hour, or
* The number `0`, which will prevent any changesets from being reconciled when this window is active.

#### `days`

`days` is an array of strings that defines the days of the week that the window applies to. English day names are accepted in a case insensitive manner:

* `["saturday", "sunday"]` constrains the window to Saturday and Sunday.
* `["tuesday"]` constrains the window to only Tuesday.

If omitted or an empty array, all days of the week will be matched.

#### `start` and `end`

`start` and `end` define the start and end of the window on each day that is matched by [`days`](#days), or every day of the week if `days` is omitted. Values are defined as `HH:MM` in UTC.

Both `start` and `end` must be provided or omitted: providing only one is invalid.

### Examples

To rate limit changeset publication to 3 per minute between 08:00 and 16:00 UTC on weekdays, and allow unlimited changesets outside of those hours:

```json
[
  {
    "rate": "unlimited"
  },
  {
    "rate": "3/minute",
    "days": ["monday", "tuesday", "wednesday", "thursday", "friday"],
    "start": "08:00",
    "end": "16:00"
  }
]
```

To only allow changesets to be reconciled at 1 changeset per minute on (UTC) weekends:

```json
[
  {
    "rate": "1/minute",
    "days": ["saturday", "sunday"]
  }
]
```

## Incoming webhooks

> NOTE: This feature was added in Sourcegraph 3.33.

Sourcegraph can track incoming webhooks from code hosts to more easily debug issues with webhook delivery. These webhooks can be viewed by going to **Site Admin > Batch Changes > Incoming webhooks**.

By default, sites without [database encryption](encryption.md) enabled will retain three days of webhook logs. Sites with encryption will not retain webhook logs by default, as webhooks may include sensitive information; these sites can enable webhook logging and optionally configure encryption for them by using the settings below.

### Enabling webhook logging

Webhook logging is controlled by the `webhook.logging` site configuration
option. This option is an object with the following keys:

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `enabled` | `boolean` | If `true`, incoming webhooks will be stored. | `true` if no site encryption is enabled; `false` otherwise. |
| `retention` | `string` | The length of time to retain the webhooks, expressed as a valid [Go duration](https://pkg.go.dev/time#ParseDuration). | `72h` |

#### Examples

To disable webhook logging:

```json
{
  "webhook.logging": {"enabled": false}
}
```

To retain webhook logs for one day:

```json
{
  "webhook.logging": {
    "enabled": false,
    "retention": "24h"
  }
}
```

### Encrypting webhook logs

Webhook logs can be encrypted by specifying a `webhookLogKey` in the [on-disk database encryption site configuration](encryption.md).
