## Нагрузочное тестирование

### Результаты:
#### Набор данных - 20 команд, 200 пользователей, 2000 пулл реквестов.

##### Для постоянного RPS = 5 100-ый перцентиль времени ответа равен ~20ms.
```bash
     scenarios: (100.00%) 1 scenario, 10 max VUs, 1m30s max duration (incl. graceful stop):
              * constant_request_rate: 5.00 iterations/s for 1m0s (maxVUs: 10, gracefulStop: 30s)

  █ THRESHOLDS 

    http_req_duration
    ✓ 'p(100)<300' p(100)=19.79ms


  █ TOTAL RESULTS 

    checks_total.......: 1287   21.44964/s
    checks_succeeded...: 98.36% 1266 out of 1287
    checks_failed......: 1.63%  21 out of 1287

    HTTP
    http_req_duration..............: avg=3.85ms min=535.96µs med=1.99ms max=19.79ms p(90)=8.66ms  p(95)=14.92ms
      { expected_response:true }...: avg=4.04ms min=535.96µs med=2.2ms  max=19.79ms p(90)=8.82ms  p(95)=15.2ms 
    http_req_failed................: 6.38%  21 out of 329
    http_reqs......................: 329    5.483241/s

    EXECUTION
    iteration_duration.............: avg=5.19ms min=76.19µs  med=3.45ms max=22.65ms p(90)=10.08ms p(95)=15.86ms
    iterations.....................: 300    4.999916/s
    vus............................: 0      min=0         max=0 
    vus_max........................: 10     min=10        max=10

    NETWORK
    data_received..................: 1.0 MB 17 kB/s
    data_sent......................: 62 kB  1.0 kB/s

running (1m00.0s), 00/10 VUs, 300 complete and 0 interrupted iterations
constant_request_rate ✓ [======================================] 00/10 VUs  1m0s  5.00 iters/s
```

#### Для постоянного RPS = 1550 100-ый перцентиль времени ответа равен ~216ms. При большем RPS начинается резкая деградация
```bash
     scenarios: (100.00%) 1 scenario, 4000 max VUs, 1m30s max duration (incl. graceful stop):
              * constant_request_rate: 1550.00 iterations/s for 1m0s (maxVUs: 4000, gracefulStop: 30s)

  █ THRESHOLDS 

    http_req_duration
    ✓ 'p(100)<300' p(100)=216.55ms


  █ TOTAL RESULTS 

    checks_total.......: 400595 6673.237868/s
    checks_succeeded...: 98.90% 396206 out of 400595
    checks_failed......: 1.09%  4389 out of 400595

    HTTP
    http_req_duration..............: avg=15.25ms min=286.97µs med=1.92ms max=216.55ms p(90)=45.88ms p(95)=106.17ms
      { expected_response:true }...: avg=15.4ms  min=286.97µs med=2ms    max=216.55ms p(90)=46.11ms p(95)=106.19ms
    http_req_failed................: 4.87%  4876 out of 99993
    http_reqs......................: 99993  1665.714934/s

    EXECUTION
    iteration_duration.............: avg=17.21ms min=24.23µs  med=2.79ms max=354.19ms p(90)=49.81ms p(95)=111.12ms
    iterations.....................: 93001  1549.239993/s
    vus............................: 76     min=2             max=272 
    vus_max........................: 4000   min=4000          max=4000

    NETWORK
    data_received..................: 614 MB 10 MB/s
    data_sent......................: 19 MB  311 kB/s

running (1m00.0s), 0000/4000 VUs, 93001 complete and 0 interrupted iterations
constant_request_rate ✓ [======================================] 0000/4000 VUs  1m0s  1550.00 iters/s
```

#### Кратко об окружении

Заклинание `make all`:
1. пересоздаёт БД
2. генерирует тестовые данные в формате `.csv`
3. из `.csv` тестовые данные загружаются в бд
4. некоторые тестовые данные копируются в файлы `.data.json` (чтобы потом использовать при тестировании)

##### Запуск
Сервис должен быть доступен на `localhost:8080`
```bash
make
k6 main_load_test.js
```
