from server import Server
import sequences


def generate_new_user():
    return sequences.get_name(), sequences.get_pass()


def test_auth():
    url = "/api/auth"
    user, passw = generate_new_user()
    data = {"username": user, "password": passw}

    server = Server()
    response = server.post(
        endpoint=url,
        data=data
    )
    assert response.status_code == 200


def test_bad_pass():
    url = "/api/auth"
    user, passw = generate_new_user()
    data = {"username": user, "password": passw}

    server = Server()
    # сначала заведем пользователя
    response = server.post(
        endpoint=url,
        data=data
    )
    assert response.status_code == 200

    # попробуем авторизоваться с другим паролем
    data = {"username": user, "password": "bad_pass"}
    response = server.post(
        endpoint=url,
        data=data
    )
    assert response.status_code == 401


def test_get_api_info():
    url = "/api/auth"
    user, passw = generate_new_user()
    data = {"username": user, "password": passw}

    server = Server()
    response = server.post(
        endpoint=url,
        data=data
    )
    assert response.status_code == 200

    # получили токен, он нужен для подключения
    url = "/api/info"
    token = response.json()["token"]
    headers = {
        "Authorization": f"Bearer {token}"
    }
    response = server.get(
        endpoint=url,
        headers=headers
    )
    # проверим, что у вновь созданного юзера 1000 коинов
    assert response.status_code == 200
    assert 1000 == response.json()["coins"]


def test_send_coins():
    url = "/api/auth"
    user1, passw1 = generate_new_user()
    data1 = {"username": user1, "password": passw1}
    user2, passw2 = generate_new_user()
    data2 = {"username": user2, "password": passw2}
    amount = 100

    # создадим первого юзера
    server = Server()
    response = server.post(
        endpoint=url,
        data=data1
    )
    assert response.status_code == 200
    token1 = response.json()["token"]
    headers1 = {
        "Authorization": f"Bearer {token1}"
    }

    # создадим второго юзера
    server = Server()
    response = server.post(
        endpoint=url,
        data=data2
    )
    assert response.status_code == 200
    token2 = response.json()["token"]
    headers2 = {
        "Authorization": f"Bearer {token2}"
    }

    # узнаем количество денег у первого юзера
    url = "/api/info"
    response = server.get(
        endpoint=url,
        headers=headers1
    )
    assert response.status_code == 200
    oldCoins1 = response.json()["coins"]

    # узнаем количество денег у второго юзера
    url = "/api/info"
    response = server.get(
        endpoint=url,
        headers=headers2
    )
    assert response.status_code == 200
    oldCoins2= response.json()["coins"]

    # переведем деньги от первого второму
    headers = {
        "Authorization": f"Bearer {token1}"
    }
    url = "/api/sendCoin"
    coins_data = {"toUser": user2, "amount": amount}
    response = server.post(
        endpoint=url,
        data=coins_data,
        headers=headers
    )
    assert response.status_code == 200

     # узнаем количество денег у первого юзера после перевода
    url = "/api/info"
    response = server.get(
        endpoint=url,
        headers=headers1
    )
    newCoins1 = response.json()["coins"]

    # узнаем количество денег у второго юзера после перевода
    url = "/api/info"
    response = server.get(
        endpoint=url,
        headers=headers2
    )
    newCoins2= response.json()["coins"]

    # смотрим, что у первого уменьшилось на amount
    assert oldCoins1 - newCoins1 == amount

    # смотрим, что у второго увеличилось на amount
    assert newCoins2 - oldCoins1 == amount


def test_self_send_coins():
    url = "/api/auth"
    user, passw = generate_new_user()
    data = {"username": user, "password": passw}

    # создадим первого юзера
    server = Server()
    response = server.post(
        endpoint=url,
        data=data
    )
    assert response.status_code == 200
    token = response.json()["token"]
    headers = {
        "Authorization": f"Bearer {token}"
    }

    # переведем деньги самому себе
    url = "/api/sendCoin"
    coins_data = {"toUser": user, "amount": 100}
    response = server.post(
        endpoint=url,
        data=coins_data,
        headers=headers
    )
    assert response.status_code == 400


def test_send_coins_all_balance():
    url = "/api/auth"
    user1, passw1 = generate_new_user()
    data1 = {"username": user1, "password": passw1}
    user2, passw2 = generate_new_user()
    data2 = {"username": user2, "password": passw2}
    amount = 100

    # создадим первого юзера
    server = Server()
    response = server.post(
        endpoint=url,
        data=data1
    )
    assert response.status_code == 200
    token1 = response.json()["token"]

    # создадим второго юзера
    server = Server()
    response = server.post(
        endpoint=url,
        data=data2
    )
    assert response.status_code == 200

    # пытаемся перевести больше, чем есть на счёте
    headers = {
        "Authorization": f"Bearer {token1}"
    }
    url = "/api/sendCoin"
    coins_data = {"toUser": user2, "amount": 1100}
    response = server.post(
        endpoint=url,
        data=coins_data,
        headers=headers
    )
    assert response.status_code == 400

    # переведем всё
    coins_data = {"toUser": user2, "amount": 1000}
    response = server.post(
        endpoint=url,
        data=coins_data,
        headers=headers
    )
    assert response.status_code == 200

    # попытаемся перевести еще
    coins_data = {"toUser": user2, "amount": 1}
    response = server.post(
        endpoint=url,
        data=coins_data,
        headers=headers
    )
    assert response.status_code == 400


def test_send_coins_check_transactions_history():
    url = "/api/auth"
    user1, passw1 = generate_new_user()
    data1 = {"username": user1, "password": passw1}
    user2, passw2 = generate_new_user()
    data2 = {"username": user2, "password": passw2}
    amount = 100

    # создадим первого юзера
    server = Server()
    response = server.post(
        endpoint=url,
        data=data1
    )
    assert response.status_code == 200
    token1 = response.json()["token"]
    headers1 = {
        "Authorization": f"Bearer {token1}"
    }

    # создадим второго юзера
    server = Server()
    response = server.post(
        endpoint=url,
        data=data2
    )
    assert response.status_code == 200
    token2 = response.json()["token"]
    headers2 = {
        "Authorization": f"Bearer {token2}"
    }

    # переведем деньги от первого второму
    headers = {
        "Authorization": f"Bearer {token1}"
    }
    url = "/api/sendCoin"
    coins_data = {"toUser": user2, "amount": amount}
    response = server.post(
        endpoint=url,
        data=coins_data,
        headers=headers
    )
    assert response.status_code == 200

    # проверим историю переводов первого юзера
    url = "/api/info"
    response = server.get(
        endpoint=url,
        headers=headers1
    )
    assert response.status_code == 200

    user1_history = {"sent": [{"amount": amount, "toUser": user2}]}
    assert response.json()["coinHistory"] == user1_history

    # проверим историю переводов второго юзера
    response = server.get(
        endpoint=url,
        headers=headers2
    )
    assert response.status_code == 200

    user2_history = {"received": [{"amount": amount, "fromUser": user1}]}
    assert response.json()["coinHistory"] == user2_history


def test_api_buy_item():
    url = "/api/auth"
    user, passw = generate_new_user()
    data = {"username": user, "password": passw}

    # создадим юзера
    server = Server()
    response = server.post(
        endpoint=url,
        data=data
    )
    assert response.status_code == 200
    token = response.json()["token"]
    headers = {
        "Authorization": f"Bearer {token}"
    }

    # покупаем футболку
    url = "/api/buy/t-shirt"
    response = server.get(
        endpoint=url,
        headers=headers,
    )
    assert response.status_code == 200

    # еще раз купим футболку
    response = server.get(
        endpoint=url,
        headers=headers,
    )
    assert response.status_code == 200

    # купим чашку
    url = "/api/buy/cup"
    response = server.get(
        endpoint=url,
        headers=headers,
    )
    assert response.status_code == 200

    # попробуем купить что-то несуществующее
    url = "/api/buy/undefined"
    response = server.get(
        endpoint=url,
        headers=headers,
    )
    assert response.status_code == 500

    # проверим историю покупок
    url = "/api/info"
    response = server.get(
        endpoint=url,
        headers=headers
    )
    assert response.status_code == 200
    inventory = [{"quantity": 2, "type": "t-shirt"}, {"quantity": 1, "type": "cup"}]
    assert inventory == response.json()["inventory"]

    # получим количество денег у юзера, создадим
    # еще одного, переведем ему остаток и попробуем
    # что-нибудь купить
    coins = response.json()["coins"]

    url = "/api/auth"
    user2, passw2 = generate_new_user()
    data2 = {"username": user2, "password": passw2}
    response = server.post(
        endpoint=url,
        data=data2
    )
    assert response.status_code == 200

    url = "/api/sendCoin"
    coins_data = {"toUser": user2, "amount": coins}
    response = server.post(
        endpoint=url,
        data=coins_data,
        headers=headers
    )
    assert response.status_code == 200

    # пытаемся купить чашку
    url = "/api/buy/cup"
    response = server.get(
        endpoint=url,
        headers=headers,
    )
    assert response.status_code == 400
