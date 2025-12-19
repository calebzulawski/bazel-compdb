#include "example/src/greeter.h"

#include <sstream>

std::string MakeGreeting(const std::string& name) {
  std::ostringstream out;
  out << "Hello, " << name << "!";
  return out.str();
}
