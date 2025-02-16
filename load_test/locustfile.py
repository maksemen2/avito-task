# Нагрузочный тест для магазина AvitoShop
# Сценарий:
# 1. Пользователь авторизуется
# 2. Пользователь случайным образом переводит монеты другому пользователю
# 3. Пользователь случайным образом покупает товар
# 4. Пользователь проверяет баланс
# 5. Если пользователь не может продолжать передавать монеты или покупать товары - он начинает спамить получением информации

# В данном тесте ошибочными считаются ответы 500, 401 и 504. При 500 и 504 коде сервер не обработал запрос как нужно, а при 401 -
# можно считать, не начал обрабатывать запрос, и он даже не попал в хендлер.

# С этим тестом сервер в контейнере держал 1726.2 RPS со средним временем ответа 47.33 мс. Процент успешности ответа - 100%
# По итогам теста количество покупок составило 9096, переводов - 18014, а пользователей - 4961.

from locust import FastHttpUser, task, between, tag
import random
import uuid
from threading import Lock
import os

goods = [
    ("t-shirt", 80),
    ("cup", 20),
    ("book", 50),
    ("pen", 10),
    ("powerbank", 200),
    ("hoody", 300),
    ("umbrella", 200),
    ("socks", 10),
    ("wallet", 50),
    ("pink-hoody", 500),
]

class SharedState:
    user_pool = []
    lock = Lock()
    user_counter = 0

class AvitoShopUser(FastHttpUser):
    wait_time = between(0.001, 0.005)
    connection_timeout = 10
    network_timeout = 10

    MAX_TRANSFERS = 2
    MAX_PURCHASES = 1
    INITIAL_BALANCE = 1000

    def on_start(self):
        with SharedState.lock:
            SharedState.user_counter += 1
            self.username = f"user{SharedState.user_counter}"
            SharedState.user_pool.append(self.username)
        
        with self.client.post("/api/auth", 
            json={"username": self.username, "password": "password"},
            catch_response=True
        ) as response:
            if response.status_code == 500:
                response.failure("Auth 500 error")
                self.stop()
            else:
                response.success()
                if response.status_code == 200:
                    self.token = response.json()["token"]
                    self.balance = self.INITIAL_BALANCE
                    self.transfer_count = 0
                    self.purchase_count = 0
                else:
                    self.stop()

    def _get_headers(self):
        return {
            "Authorization": f"Bearer {self.token}",
        }

    @task(5)
    @tag("transfer")
    def transfer_coins(self):
        if self.transfer_count >= self.MAX_TRANSFERS:
            return
            
        if len(SharedState.user_pool) > 1:
            recipient = random.choice([
                u for u in SharedState.user_pool 
                if u != self.username
            ])
        else:
            return

        if self.balance >= 100:
            with self.client.post("/api/sendCoin",
                json={"toUser": recipient, "amount": 100},
                headers=self._get_headers(),
                catch_response=True
            ) as response:
                if response.status_code == 500  or response.status_code == 401 or response.status_code == 504:
                    response.failure("Transfer error")
                else:
                    response.success()
                
                if response.status_code == 200:
                    self.transfer_count += 1
                    self.balance -= 100

    @task(8)
    @tag("purchase")
    def purchase_item(self):
        if self.purchase_count >= self.MAX_PURCHASES:
            return

        affordable = [(item, price) for item, price in goods if price <= self.balance]
        if not affordable:
            return

        item, price = random.choice(affordable)
        with self.client.get(f"/api/buy/{item}",
            headers=self._get_headers(),
            catch_response=True,
            name=f"/api/buy/ [cached]"
        ) as response:
            if response.status_code == 500  or response.status_code == 401 or response.status_code == 504:
                response.failure("Purchase error")
            else:
                response.success()
            
            if response.status_code == 200:
                self.purchase_count += 1
                self.balance -= price

    @task(1)
    def check_balance(self):
        with self.client.get("/api/info", 
            headers=self._get_headers(),
            catch_response=True
        ) as response:
            if response.status_code == 500  or response.status_code == 401 or response.status_code == 504:
                response.failure("Check balance 500 error")
            else:
                response.success()

    @task(2)
    def mixed_operation(self):
        self.transfer_coins()
        self.purchase_item()