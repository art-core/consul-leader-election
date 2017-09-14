# Consul Leader Election CHANGELOG

## 0.3.0

##### Fixed:
  * if acquiring the session is not successful, exit with `-not-leader-exit-code` instead of `0`.

##### Added:
  * `-service-name`: Name of the service you want to tag.
  * `-leader-tag`: Tag which will be set to `-service-name` if leader.
  * `-not-leader-tag`: Tag which will be set to `-service-name` if not leader.
  * `-session-lock-delay`: The session's lock-delay time in seconds. (Default: 1)

## 0.2.2

##### Fixed:
  * check if key exists before checking key session

## 0.2.1

##### Fixed:
  * acquire session if key exists but has no session

## 0.2.0

##### Added:
  * `-leader-exit-code`: overwrite exit code if leader (Default: 0)
  * `-not-leader-exit-code`: overwrite exit code if not leader (Default: 1)
  * `-error-exit-code`: overwrite exit code for errors (Default: 2)

##### Changed:
  * check key before trying to acquire it

## 0.1.0

  * Initial release
