# Day 14: Lost in Nairobi CBD

## The brief

It's your first week in Nairobi. You got off the bus at Archives, it's 6 p.m., and your mat home leaves from Kencom. Every makanga is shouting a different route and Google Maps has given up. Build a text adventure about it: the player types commands in sheng, walks from street to street, and wins by reaching Kencom.

This is the Phase 1 capstone. There is no walkthrough today. You get a brief, a list of requirements, the tests and some hints. Everything you need is from Days 1 to 13.

```
Step("archives", "nenda kusini")     →  "hilton", "Uko Hilton. Watu wa ofisi wanatembea fast na simu kwa sikio.\nNjia: kaskazini, kusini, magharibi"
Step("archives", "N")                →  "odeon", "Uko Odeon. ...\nNjia: kusini, mashariki"
Step("kimathi", "nenda kaskazini")   →  "kimathi", "Huwezi enda huko. Hakuna njia."
Step("archives", "kula smokie")      →  "archives", "Sijakuelewa, msee. Jaribu 'saidia'."
Step("hilton", "s")                  →  "kencom", "Uko Kencom! ...\nNjia: kaskazini\nUmefika! Panda mat uende home. Umeshinda!"
```

A game in a terminal:

```
Uko nje ya Archives kwa Moi Avenue. Makanga wanapiga kelele, "Rongai, Rongai!" Hii si stage yako.
Njia: kaskazini, kusini, mashariki
> nenda kusini
Uko Hilton. Watu wa ofisi wanatembea fast na simu kwa sikio.
Njia: kaskazini, kusini, magharibi
> s
Uko Kencom! Mat yako ya kwenda home iko hapa.
Njia: kaskazini
Umefika! Panda mat uende home. Umeshinda!
```

The whole map is data, already written for you in `starter/adventure.go`:

```go
var descriptions = map[string]string{"archives": "Uko nje ya Archives ...", ...}
var exits = map[string]map[string]string{
	"archives": {"kaskazini": "odeon", "kusini": "hilton", "mashariki": "river-road"},
	...
}
```

`exits["archives"]["kusini"]` is `"hilton"`: walk south from Archives and you reach Hilton. A room is just its name, a string. Day 15 introduces structs, which are a neater way to group a room's description and exits. You don't need one today: two maps that share the same keys do the job.

## What you'll use

- Maps of maps, and the comma-ok idiom, to look up an exit that might not exist.
- `strings.Fields` and `strings.ToLower` to read a command whatever the spacing and case.
- `sort.Strings` and `strings.Join` to list the exits in a fixed order.
- A function that takes the current state and an input and returns the new state and the output. That one idea makes a whole game testable without typing a single command.
- `bufio.Scanner` on `os.Stdin` for the playable version, as on Day 12.

```go
next, ok := exits["kimathi"]["kaskazini"] // "", false: there's no road that way

ways := []string{"kusini", "kaskazini"}
sort.Strings(ways)
strings.Join(ways, ", ") // "kaskazini, kusini"
```

## Steps

Requirements. Meet them in any order you like.

1. `Describe(room)` returns the room's description, a newline, then `Njia: ` and the room's exits in alphabetical order, separated by `, `. An unknown room gives `""`.
2. `Won(room)` is true only for `Goal`.
3. `Step(room, command)` ignores case and extra spaces. It returns the new room and the text to show.
4. `angalia` describes the room again. `saidia` returns `Help`. Neither moves the player.
5. `nenda <njia>` moves the player. A direction on its own works too, so `kusini` and `nenda kusini` do the same thing. The four directions are `kaskazini`, `kusini`, `mashariki` and `magharibi`, and `n`, `s`, `e` and `w` are short for them.
6. After a move, the text is `Describe` of the new room. On reaching `Goal`, add `\nUmefika! Panda mat uende home. Umeshinda!` to the end.
7. Every other case keeps the player where they are, with exactly one of these messages:
   - empty command: `Sema kitu, msee.`
   - `nenda` with no direction: `Nenda wapi? Sema njia.`
   - `nenda` with a word that isn't a direction: `Hiyo si njia. Jaribu kaskazini, kusini, mashariki au magharibi.`
   - a direction with no exit that way: `Huwezi enda huko. Hakuna njia.`
   - an unknown command: `Sijakuelewa, msee. Jaribu 'saidia'.`
   - a room that isn't on the map: `Hiyo place sijui.`
8. Write the game loop in `starter/cmd/play/main.go`. It prints `> `, reads a line, stops on `toka` or the end of input, calls `Step` and prints the text, until the player wins.
9. Run the tests from the repo root, then play it:

```bash
go test ./phase-1-terminal/day-014-lost-in-nairobi/starter
go run ./phase-1-terminal/day-014-lost-in-nairobi/starter/cmd/play
```

When it works, make it yours: add a room for Tom Mboya Street or Luthuli Avenue to both maps. Add its exits, and an exit from a neighbouring room that leads to it. The tests will still pass as long as the rooms they check don't change.

## The kata: Grid Walk

A delivery rider walks a grid of CBD blocks, following a list of moves. Walls stop them. Print where they end up. Grid simulation is a classic HackerRank problem, and it is your game's map turned into a 2D array. Instructions are in `kata/README.md`. Judge it with:

```bash
go run ./tools/kata 14
```

## When you're stuck

- **`assignment to entry in nil map`**: you wrote into a map you declared with `var m map[string]int`. Make it with `map[string]int{}` or `make` first. Reading a nil map is fine; writing panics.
- **`Describe` lists the exits in a different order each run**: map iteration order is random. Collect the keys into a slice and `sort.Strings` it.
- **`index out of range [0] with length 0`**: an empty command has no words. Check `len(words) == 0` before you read `words[0]`.
- **`nenda N` doesn't move**: lower-case the whole command before splitting it, so `N` becomes `n`, then turn short directions into full ones with a small map.
- **The player ends up in room `""`**: you used `exits[room][way]` without checking `ok`. A missing key returns `""`, so return the old room when `ok` is false.
- **`declared and not used: text`** in your loop: print the text that `Step` returns. Use `room, text = Step(room, line)` with `=`, not `:=`, so you update the `room` outside the loop.
- **Still stuck?** Open `solution/adventure.go` and `solution/cmd/play/main.go`, read them once, close them, then write your own version from memory.
