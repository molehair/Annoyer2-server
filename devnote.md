## History

- 

## Device sync

- Add
  - When device B is added to device A, B.userCode is replaced by A.userCode.
  - All words in A, B, and the server must be sync then.
- Delete
  - When device B is detached from device A, B.userCode is generated.
- Must consider that add and delete operations may happen over more than two devices simultaneously.



## Word sync

- add
- modify
- delete

## Push notification

Use FCM

## DB

- dictionary
  - *userCode*_ *syncId* : {word:_ , def:_ , ... }
  - {...} doesn't include wordId, syncId and userCode
- users
  - *userCode* : {deviceIds: [ .. ]}
- devices
  - *deviceId*_ fcmToken : {token:_ , timestamp: _ }
  - *deviceId*_ trainingSchedule :  *hhmmwwwwwww*
  - timestamp is for maintaining the token as recent one
- trainingSchedules
  - *hhmmwwwwwww*_ *deviceId* : "" (empty string)

#### terms

- hhmmwwwwwww
  - hh in [00, ... , 59] meaning hour
  - mm in [00, ... , 59] meaning minute
  - In wwwwwww, a *w* is either 0 or 1, meaning the alarm is disabled/enabled on Sun, ... , Sat, respectively.
  - All values are UTC time.

## Todo

- Ignore old request by saving the latest timestamp in the packet
  - ex) Changing training time may incur multiple packets, but the only recent one is required.
  - The timestamp must be generated and sent by client.
  - Save the timestamp in the user's session.

## Build docker image

This instruction is written based on `Docker version 20.10.14, build a224086`, targeting cross-architecture.

1. Check supported architectures.

```bash
docker buildx ls
```

If the architecture you want is not shown, install qemu on Linux or Docker desktop on Windows and Mac.

2. Build

```bash
docker buildx build --platform linux/amd64,linux/arm64,linux/arm/v7 -t molehair/annoyer2server:latest . --push
```