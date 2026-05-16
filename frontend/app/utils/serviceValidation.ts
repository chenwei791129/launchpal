// Mirrors the launchd plist invariant enforced by the backend
// (internal/launchctl/user.go validateProgramOrArguments): a service config
// MUST specify either Program or at least one entry in Arguments. Both empty
// would write a plist that launchd refuses to load.
export function hasProgramOrArguments(program: string, argumentsText: string): boolean {
  return program !== '' || argumentsText.trim() !== ''
}

export const PROGRAM_PATH_HINT =
  'Optional. Leave empty if the executable is provided as the first item in Arguments.'
