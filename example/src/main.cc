#include <iostream>

#include "example/src/greeter.h"

int main(int argc, char** argv) {
  const char* who = argc > 1 ? argv[1] : "World";
  std::cout << MakeGreeting(who) << std::endl;
  return 0;
}
