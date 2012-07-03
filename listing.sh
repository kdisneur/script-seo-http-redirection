#/bin/sh
function usage {
  printf " usage: $0 <INPUT_FILE> <OUTPUT_FILE> [<CHUNK_SIZE>]\n"
  printf " exple: $0 ~/Desktop/liens-casses.txt ~/Desktop/results.csv \n"
}

function check_params {
  [ -f "$1" ] || {   printf "La commande attend un fichier d'urls en paramètre!\n"; usage; exit 1; }
  [ -n "$2" -a ! -e $2 ] || {   printf "Le fichier passé en paramètre existe déjà!\n"; usage; exit 1; }
  [ -n "$3" ] || { export chunk_size=10; printf "default chunk size used: ${chunk_size}\n"; }
}

function write {
  url=$1
  status=$2
  destination=$3
  printf "%s;%s;%s\n" ${url} ${status} ${destination} >> ${output_file}
}

function same_http_code {
  base_status=$1
  status=$2
  return_code=1
  if [ "${base_status}" = "${status}" ]; then
    return_code=0
  fi
  return ${return_code}
}

function redirected_status {
  status=$1
  http_redirect_status="301 302 307"
  return $(grep -E "(^| )${status}( |$)" > /dev/null <<< ${http_redirect_status})
}

function analyze_urls {
  input_file=$1
  local url_analyzed=0
  for url in $(cat ${input_file}); do
    analyze_url ${url} &
    url_analyzed=$((${url_analyzed} + 1))
    if [ ${url_analyzed} -eq ${chunk_size} ]; then
      wait
      url_analyzed=0
    fi
  done
  wait
}

function analyze_url {
  local url=$1
  local status=$(curl -I ${url} 2> /dev/null | grep "HTTP/1.1" | sed -E "s/HTTP\/1.1 ([0-9]+).*/\1/")
  if redirected_status ${status}; then
    local destination=$(curl -I ${url} 2> /dev/null | grep "Location:" | sed -E "s/Location: (.*)/\1/")
  fi
  write "${url}" "${status}" "${destination}"
}

function main {
  export input_file=$1
  export output_file=$2
  export chunk_size=$3
  check_params ${input_file} ${output_file} ${chunk_size}
  analyze_urls ${input_file}
}

main $*
