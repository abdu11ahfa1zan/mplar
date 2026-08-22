import curses
from curses import wrapper

def main(stdscr):
    pass
    stdscr.clear()
    stdscr.refresh()
    stdscr.addstr("Type", curses.A_UNDERLINE)
    stdscr.getch()
wrapper(main)