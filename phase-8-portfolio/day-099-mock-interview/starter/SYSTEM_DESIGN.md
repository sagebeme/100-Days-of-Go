# System design: Twende Seats

> Design a booking system for a matatu SACCO with 200 vehicles. Passengers book a
> seat on a trip from a smartphone app or by USSD on a feature phone, pay with an
> M-Pesa STK push, and get their ticket by SMS. Normal days bring about 50 000
> bookings; the week before Christmas brings ten times that. Two people must never
> get the same seat.

Time: 45 minutes. Spend about 5 on questions, 5 on numbers, 20 on the design, 15 on
the deep dives. Write in short bullets; draw boxes and arrows in ASCII or on paper.

## 1. Questions I'd ask, and the answers I'm assuming

-

## 2. Requirements

**Functional (what it does)**

-

**Non-functional (how well: latency, availability, consistency, cost)**

-

## 3. Numbers

<!-- Bookings per second on average, at the daily peak, and in December. Data per year. -->

-

## 4. API

<!-- The main endpoints: method, path, request, response. -->

-

## 5. Data model

<!-- Tables, key columns, and the constraint that stops double booking. -->

-

## 6. High-level design

<!-- The boxes and arrows, and one sentence on what each box does. -->

```
```

## 7. Deep dives

**No double booking:**

-

**The payment callback (late, lost, or sent twice):**

-

## 8. What breaks, and how it scales

-

## 9. Trade-offs I made

-
