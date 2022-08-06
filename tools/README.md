# overflow-genearte

I use  these mappings in nvim

nmap <leader>c /<><cr>xxi
inoremap <silent><c-q> <esc>/<><cr>xxi

so I can run
> o scripts/<filename>.cdc

it generates a template from that and opens it in your favorite editor nvim, 
You press <leader>c and are taken to the first placeholder, you fix that and in insert mode press c-q and are taken to the next one. 

Once you exit nvim it will perform the interaction. 

## building
build overflow-generate and put it in your path


## disclaimer
This has only been tested on my mac. 


## improvements

# overflow-json

build it, and use with ´query´ script to get sane json output from flow-cli
