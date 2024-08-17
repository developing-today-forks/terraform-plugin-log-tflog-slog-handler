Until upstream terraform-plugin-loh-tflog has a slog alternative

https://github.com/hashicorp/terraform-plugin-log/issues/108

A separate package because I don't want to fork a provider.

This way one can pull this Handler in without overriding hashicorp's repo or having a delay before updates.

Use Case:

A dependency accepts slog Handler or uses default slog.Handler.
