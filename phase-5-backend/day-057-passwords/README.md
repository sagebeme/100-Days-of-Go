# Day 57: Passwords

## The brief

Twende is a (made-up) matatu SACCO app. Passengers sign up with a phone number and a password, then book seats. The old system was written in a hurry: it stored bcrypt hashes at a low cost, and an intern once added a log line that printed the password on every failed login. Today you write the password code properly.

The rules: a password is checked for length, not for "one capital, one digit, one symbol". It is stored only as an argon2id hash. Old bcrypt hashes still work, and each one is quietly replaced with argon2id the next time its owner logs in. Nothing secret ever reaches the log.

```
ValidatePassword("matatu-ya-saa-6", blocklist)          →  nil
ValidatePassword("short", blocklist)                    →  ErrTooShort
ValidatePassword("PasswordPassword", blocklist)         →  ErrCommon

HashArgon2id("ngong road ni jam tupu", DefaultParams)
  →  "$argon2id$v=19$m=19456,t=2,p=1$Vc2t...$3bQ8..."   (different every time)

Verify(argon2idHash, "ngong road ni jam tupu", DefaultParams)  →  ok=true,  rehash=false
Verify(oldBcryptHash, "ngong road ni jam tupu", DefaultParams) →  ok=true,  rehash=true
Verify(argon2idHash, "wrong guess", DefaultParams)             →  ok=false, rehash=false

svc.Login("0712345678", "ngong road ni jam tupu")  →  nil
svc.Login("0712345678", "wrong guess")             →  ErrInvalidCredentials
svc.Login("0799999999", "anything")                →  ErrInvalidCredentials   (same answer)
```

## What you'll use

- **A password hash** is slow on purpose. SHA-256 runs billions of times a second on a gaming GPU, so a leaked table of SHA-256 passwords falls in hours. bcrypt and argon2id let you choose how much work (and, for argon2id, how much memory) each guess costs.
- **A salt**: 16 random bytes from `crypto/rand`, made fresh for each password and stored next to the hash. Two passengers with the same password get different hashes, and an attacker can't use one precomputed table for everyone.
- **`golang.org/x/crypto/argon2`**: `argon2.IDKey(password, salt, time, memory, threads, keyLen)`. Memory is in KiB. The OWASP Password Storage Cheat Sheet's minimum is 19 MiB, 2 passes, 1 thread, which is `DefaultParams`.
- **`golang.org/x/crypto/bcrypt`**: `GenerateFromPassword(pw, cost)` and `CompareHashAndPassword(hash, pw)`. The cost is a power of two: each step up doubles the work. `bcrypt.Cost(hash)` reads it back. bcrypt only uses the first 72 bytes. `GenerateFromPassword` returns `bcrypt.ErrPasswordTooLong` for longer input; `CompareHashAndPassword` doesn't complain, it just looks at the first 72 bytes.
- **The PHC string format** packs everything needed to check a password into one string, so you can change the settings later without losing old hashes:

```
$argon2id$v=19$m=19456,t=2,p=1$<salt in base64>$<hash in base64>
```

- **Constant-time comparison**: `subtle.ConstantTimeCompare(a, b)` from `crypto/subtle` takes the same time whether the first byte or the last byte differs. A plain `bytes.Equal` stops at the first difference, and that timing can leak how close a guess was.
- **`unicode/utf8.RuneCountInString`** counts characters, not bytes, so `"é"` is 1 character even though it is 2 bytes.

```go
salt := make([]byte, 16)
rand.Read(salt) // crypto/rand, never math/rand
key := argon2.IDKey([]byte(pw), salt, 2, 19*1024, 1, 32)
b64 := base64.RawStdEncoding // standard alphabet, no "=" padding
s := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, 19*1024, 2, 1,
	b64.EncodeToString(salt), b64.EncodeToString(key))

if subtle.ConstantTimeCompare(got, want) == 1 { /* match */ }
```

### The password rules: NIST SP 800-63B

The US standards body NIST publishes the rules most security teams follow, in [SP 800-63B](https://pages.nist.gov/800-63-4/sp800-63b.html). The current revision (800-63B-4, 2025) says:

- At least **15 characters** when the password is the only thing a user proves (no SMS code or app as a second factor). At least 8 is allowed only as part of multi-factor login.
- Allow at least **64 characters**, and allow spaces and any Unicode. Count each Unicode character as one.
- **No composition rules.** Don't demand a digit, a capital or a symbol. They make people write `Password1!`, which is in every cracking list.
- **Check against a blocklist** of common and breached passwords, and say why you rejected it.
- Don't force people to change passwords on a schedule. Only when there's a sign it leaked.

Length beats complexity: `nimechoka na jam ya thika road` is easier to remember and far harder to guess than `Tw3nd3!`.

## Why it matters at work

Password storage is one of the first things a security reviewer checks, because the database will leak one day: a backup on a shared drive, a SQL injection, a laptop left in a matatu. When it does, the hash is all that stands between the attacker and every passenger's password, which many of them reuse for M-Pesa and email. In code review you'll be asked: is it argon2id or bcrypt with a reasonable cost? Is the salt random per user? Is the comparison constant-time? Does the login answer the same for an unknown phone and a wrong password? Does any log line, error message or analytics event contain the password? And when the team raises the cost next year, does anything upgrade the old hashes? Today's code answers all of those. See A02 and A07 in [owasp-checklist.md](../../reference/owasp-checklist.md).

## Steps

1. Open `starter/passwords.go`. The errors, `Params`, `DefaultParams` and the in-memory `MemStore` are written for you.
2. Write `ValidatePassword`. Count characters with `utf8.RuneCountInString`. Return `ErrTooShort` under `MinLength`, `ErrTooLong` over `MaxLength`, and `ErrCommon` if `strings.ToLower(password)` is in the blocklist. Nothing else.
3. Write `HashArgon2id`. Make a salt of `p.SaltLen` bytes with `crypto/rand`, call `argon2.IDKey`, and format the PHC string shown above with `base64.RawStdEncoding`.
4. Write `VerifyArgon2id`. Split the string on `"$"`: you should get 6 parts, the first empty and the second `argon2id`. Read the version with `fmt.Sscanf(parts[2], "v=%d", &v)` and the costs with `fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", ...)`. Decode the salt and the key. Any problem returns `ErrMalformedHash`. Recompute the key with the same salt, parameters and key length, and compare with `subtle.ConstantTimeCompare`.
5. Write `HashBcrypt` with `bcrypt.GenerateFromPassword`. Return its error as it is.
6. Write `Verify`. Look at the prefix: `$argon2id$` goes to your argon2id code, and `needsRehash` is true when the stored parameters differ from `want`. `$2a$`, `$2b$` or `$2y$` is bcrypt: use `bcrypt.CompareHashAndPassword`, and a match always needs a rehash, because new hashes are argon2id. A wrong password is `ok=false` with no error. Anything else is `ErrMalformedHash`.
7. Write `Register`: `ErrPhoneTaken` if the phone exists, then `ValidatePassword`, then hash and `Store.Put`. Log `"user registered"` with the phone only.
8. Write `Login`. For an unknown phone, still run `Verify` against a hash of some throwaway password, so the answer takes as long as for a real phone, then return `ErrInvalidCredentials`. For a known phone, `Verify`; any error or mismatch is also `ErrInvalidCredentials`. On success with `needsRehash`, hash again with `s.Params` and save it. Log `"login failed"` (with the phone and a reason) on every failure, and never log the password or the hash.
9. Run the tests from the repo root:

```bash
go test ./phase-5-backend/day-057-passwords/starter
```

The tests use tiny argon2id settings so they run fast. To feel the real cost, time one hash with `DefaultParams`: about 20 to 50 ms on a laptop. That's nothing for one login and a wall for someone trying a billion guesses.

## The kata: Account Lockout

Replay a log of login attempts and print the server's answer to each: three wrong passwords in a row lock the account for five minutes. Hashing makes each guess slow; lockout and rate limiting ([Day 61](../day-061-cors-and-security-headers/)) cap how many guesses there are. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 57
```

## When you're stuck

- **`illegal base64 data at input byte 22`**: you encoded with `base64.StdEncoding` (with `=` padding) and decoded with `RawStdEncoding`, or the other way round. Use `RawStdEncoding` for both.
- **`VerifyArgon2id` always returns false for the right password**: you recomputed with `DefaultParams` instead of the parameters read from the string, or used a new salt. Use exactly what's stored.
- **`two hashes of the same password are equal`**: the salt isn't random. Make it with `crypto/rand.Read` on every call.
- **`Verify` says rehash for a fresh argon2id hash**: compare the full `Params`, including `SaltLen` and `KeyLen`, which come from the decoded lengths of the salt and key.
- **`bcrypt: password length exceeds 72 bytes`**: bcrypt's limit. `MaxLength` is 64 characters, but 64 multi-byte characters can be over 72 bytes. That's one more reason new hashes use argon2id, which has no such limit.
- **`crypto/bcrypt: hashedPassword is not the hash of the given password`**: that's `bcrypt.ErrMismatchedHashAndPassword`, a wrong password, not a broken hash. Return `ok=false` with a nil error for it.
- **The test says the log contains the password**: look for `"password", password` or `"err", err` where the error message includes the input. Log the phone and a reason, nothing more.
- **Still stuck?** Open `solution/passwords.go`, read it once, close it, then write your own version from memory.
