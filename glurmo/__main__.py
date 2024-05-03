# pyright: strict
import sys
import os
import subprocess
from .setup import setup
from .run import run
from .cancel import cancel
from .status import status

def main():
    if len(sys.argv) < 2:
        print("Usage: glurmo [path/to/target]")
        return None

    target_dir = sys.argv[1]
    target_dir = os.path.abspath(target_dir)
    username = bytes.decode(subprocess.check_output("whoami"), encoding='utf-8')
    username = username.strip()

    print("Managing: " + target_dir)
    print("As:       " + username)

    while True:
        print("What would you like to do next?")
        print(
"""
q -> quit
s -> run setup
r -> run sims
c -> cancel runs
t -> status
"""
        )
        
        next_action = input().split()
        if next_action[0] == 'q':
            break
        elif next_action[0] == 's':
            setup(target_dir)
        elif next_action[0] == 'r':
            run(next_action, target_dir, username)
        elif next_action[0] == 'c':
            cancel(next_action, target_dir, username)
        elif next_action[0] == 't':
            status(target_dir, username)
        else:
            print("Unrecognized action")

if __name__ == '__main__':
    main()