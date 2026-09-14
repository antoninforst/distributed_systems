import sys

COLOR_GREEN = "\033[92m"
COLOR_YELLOW = "\033[93m"
COLOR_RESET = "\033[0m"

raw_text = sys.stdin.read()

lines = [l.strip() for l in raw_text.strip().splitlines() if ":" in l]
philosophers = set(l.split(":", 1)[0].strip() for l in lines)

sorted_philosophers = sorted(
    list(philosophers),
    key=lambda x: int(x[1:]) if x[1:].isdigit() else x
)

# Všichni filozofové začínají ve stavu Thinking
current_states = {p: "Thinking" for p in sorted_philosophers}
timeline = {p: [] for p in sorted_philosophers}

for line in lines:
    p_id, state = [part.strip() for part in line.split(":", 1)]
    state = state.capitalize()

    # Sloupec se přidá pouze při reálné změně stavu
    if current_states[p_id] != state:
        current_states[p_id] = state
        for p in sorted_philosophers:
            if current_states[p] == "Eating":
                timeline[p].append(f"{COLOR_GREEN}G{COLOR_RESET}")
            else:
                timeline[p].append(f"{COLOR_YELLOW}Y{COLOR_RESET}")

for p in sorted_philosophers:
    print(f"{p}: {''.join(timeline[p])}")