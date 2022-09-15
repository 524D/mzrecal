# Galaxy definition files

This folder contains the [necessary files](https://galaxyproject.org/admin/tools/add-tool-tutorial/) to use [mzRecal](https://github.com/524D/mzrecal) in [Galaxy](https://galaxyproject.org/).

To use mzRecal in Galaxy:

* In the Galaxy tools directory, create a new subdirectory `mzrecal`
* Put the mzrecal executable and `mzrecal.xml` (the tool definition file) in the newly made `mzrecal` directory
* Append the contents of `tool_conf-add.xml` to `config/tool_conf.xml`
