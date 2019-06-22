## repp make sequence

Find or build a plasmid from its target sequence

### Synopsis

Build up a plasmid from its target sequence using a combination of existing and
synthesized fragments.

Solutions have either a minimum fragment count or assembly cost (or both).

```
repp make sequence [flags]
```

### Options

```
  -a, --addgene           use the Addgene repository
  -b, --backbone string   backbone to insert the fragments into. Can either be an entry 
                          in one of the dbs or a file on the local filesystem.
  -d, --dbs string        list of local fragment databases
  -u, --dnasu             use the DNASU repository
  -e, --enzyme string     enzyme to linearize the backbone with (backbone must be specified).
                          'repp ls enzymes' prints a list of recognized enzymes.
  -x, --exclude string    keywords for excluding fragments
  -h, --help              help for sequence
  -p, --identity int      %-identity threshold (see 'blastn -help') (default 98)
  -g, --igem              use the iGEM repository
  -i, --in string         input FASTA with target sequence
  -o, --out string        output file name
```

### Options inherited from parent commands

```
  -s, --settings string   build settings (default "/Users/josh/.repp/config.yaml")
  -v, --verbose           whether to log results to stdout
```

### SEE ALSO

* [repp make](repp_make.md)	 - Make a plasmid from its fragments, features or sequence

###### Auto generated by spf13/cobra on 22-Jun-2019