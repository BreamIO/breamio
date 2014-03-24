# Bream IO
The [golang][] implementation of Eriver, aka our event software.

## Calibration protocol

|                Client | Direction | Server           |
| ------------------: | :----------: | :---------------- |
|    calibrate:add |      -->      |                       |
|                          |      <--     | calibrate:next |
|    calibrate:add |      -->      |                       |
|                          |      <--     | calibrate:next |
|    calibrate:add |      -->      |                       |
|                          |      <--     | calibrate:next |
|                          |    :zap:    |                       |
|    calibrate:add |      -->      |                       |
|                          |      <--     | calibrate:end  |
| ------------------- | ----------- | ----------------- |
|                          |     <-->    | validate:start  |
|      validate:add |      -->     |                       |
|                          |      <--     | validate:next  |
|      validate:add |      -->     |                       |
|                          |      <--     | validate:next  |
|                          |    :zap:    |                       |
|      validate:add |      -->     |                       |
|                          |      <--     | validate:end   |

[golang]: http://golang.org/
