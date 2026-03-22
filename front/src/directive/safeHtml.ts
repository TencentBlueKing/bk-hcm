import DOMPurify from 'dompurify';

export default {
  mounted(el: { innerHTML: string }, binding: { value: string | Node }) {
    el.innerHTML = DOMPurify.sanitize(binding.value);
  },
  updated(el: { innerHTML: string }, binding: { value: string | Node }) {
    el.innerHTML = DOMPurify.sanitize(binding.value);
  },
};
