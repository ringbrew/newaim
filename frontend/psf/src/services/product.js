import axios from "axios"
import { baseUrl } from "@/config"

export default function SearchProduct(from, size, keyword) {
  let code = localStorage.getItem('code')
  if(!code){
      code = GetRandomString(20);
      localStorage.setItem('code', code)
  }

  return new Promise((resolve,reject)=> { 
    axios.get(baseUrl + '/product?keyword='+keyword+'&from='+from+'&size='+size,{headers: {'X-Newaim-Api-Key': code}}).then(res=>{
      resolve(res)
    }).catch(err=>{
      reject(err)
    })
  })
}

const _charStr = 'abacdefghjklmnopqrstuvwxyzABCDEFGHJKLMNOPQRSTUVWXYZ0123456789';

function RandomIndex(min, max, i){
  let index = Math.floor(Math.random()*(max-min+1)+min),
      numStart = _charStr.length - 10;
  if(i==0&&index>=numStart){
      index = RandomIndex(min, max, i);
  }
  return index;
}


function GetRandomString(len){
  let min = 0, max = _charStr.length-1, _str = '';
  len = len || 15;
  for(var i = 0, index; i < len; i++){
      index = RandomIndex(min, max, i);
      _str += _charStr[index];
  }
  return _str;
}