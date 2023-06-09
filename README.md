# Курс обміну Bitcoin до UAH 💰

## Робота програми 🛠️

Програма працює наступним чином:
1. Вона використовує Abstract API для отримання курсу обміну Bitcoin до євро (EUR).
2. Потім вона використовує ExchangeRates API для отримання курсу обміну євро до гривні (UAH).
3. За допомогою цих двох значень, програма розраховує і повертає кінцевий курс обміну Bitcoin до гривні (UAH).
4. Користувачі можуть підписатися на оновлення курсу, вказавши свою електронну адресу.
5. Програма може надсилати повідомлення з оновленнями курсу на всі підписані електронні адреси за допомогою Gmail API.

## Примітка 📝
* 🤑 Я не зміг знайти безкоштовне API, яке безпосередньо надає курси обміну валют від BTC до UAH, тому я використав два API. Один отримує курс від BTC до EUR, а наступний - від EUR до UAH.
* 🤮 Код міг бути набагато чистішим, але через нестачу часу, і через те, що golang - це нова мова для мене, код вийшов досить "грубим". Проте, цей проект служить хорошим основним прикладом роботи з API і валютними курсами в Go.

