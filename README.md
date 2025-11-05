# Module arm-controls 

A Viam module that provides arm control services for repeating predefined movements.

## Model brad-grigsby:arm-controls:repeat-arm-movements

A generic service that controls an arm to repeat a sequence of joint positions for a specified number of iterations. This service is useful for automated testing, demonstrations, or repetitive tasks.

### Configuration
The following attribute template can be used to configure this model:

```json
{
  "arm": "<string>",
  "joint_positions": [[<float>, <float>, ...]],
  "num_repeats": <int>
}
```

#### Attributes

The following attributes are available for this model:

| Name              | Type         | Inclusion | Description                                    |
|-------------------|--------------|-----------|------------------------------------------------|
| `arm`             | string       | Required  | Name of the arm component to control           |
| `joint_positions` | [][]float64  | Required  | Array of joint position arrays to cycle through |
| `num_repeats`     | int          | Required  | Number of times to repeat the sequence (must be > 0) |

#### Example Configuration

```json
{
  "arm": "my_arm",
  "joint_positions": [
    [0.0, 0.5, 1.0, 0.0, 0.0, 0.0],
    [0.5, 0.0, 0.5, 1.0, 0.0, 0.0]
  ],
  "num_repeats": 3
}
```

### DoCommand

The service implements the following DoCommand operations:

#### Execute Sequence

Executes the full sequence of joint positions for the configured number of repeats:

```json
{
  "command": "execute"
}
```

#### Move to Specific Position

Moves the arm to a specific joint position from the configured sequence by index:

```json
{
  "command": "move_to_index",
  "index": 0
}
```

**Parameters:**
- `index` (int): Zero-based index of the joint position in the `joint_positions` array
