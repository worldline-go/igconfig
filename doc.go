/*
This package allows to set struct values based on different input places.

For example struct can be filled from default values defined in tags, environmental variables or Vault secrets.

All values in tags(except value in 'default' tag) should be lower-cased.
Failing to make values lower-case will result in inability to properly set those fields.
This is a choice to make lower number of mixed-cased values and generally be more in-line
with best practices and proper usages.


Library is basically a set of "loaders".

Loader is anything that satisfies loader.Loader interface.

Such loaders will take application name and structure that should be filled
and will do specific steps do try and set fields in structure.

LoadConfig function contains all the loaders that are enabled by default.
This function should be used in most of the cases as only functionality
that should be used by the caller.

In rare cases caller might want to use specific set of loaders to load configuration.

Some loaders also provide ability to periodically update values.
This can be useful if application has logging and caller might want to change logging level at a runtime.

// Changes!
All loaders that use internal.StructIterator will not be able to use multiple names of fields!
Env is one of them.

*/
package igconfig
